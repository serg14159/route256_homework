package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	loms_app "route256/loms/internal/app/loms"
	"route256/loms/internal/config"
	db "route256/loms/internal/pkg/db"
	internal_errors "route256/loms/internal/pkg/errors"
	kafka_producer "route256/loms/internal/pkg/kafka"
	"route256/loms/internal/pkg/shard_manager"
	repo_order "route256/loms/internal/repository/orders"
	repo_outbox "route256/loms/internal/repository/outbox"
	repo_stocks "route256/loms/internal/repository/stocks"
	"route256/loms/internal/server"
	loms_service "route256/loms/internal/service/loms"
	"route256/utils/logger"
	"route256/utils/tracer"

	oteltrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const quitChannelBufferSize = 1
const errorChannelBufferSize = 1
const processOutboxInterval time.Duration = 2 * time.Second
const stdout = "stdout"

type App struct {
	config          *config.Config
	ctx             context.Context
	cancel          context.CancelFunc
	pool            *pgxpool.Pool
	lomsService     *loms_service.LomsService
	lomsApp         *loms_app.Service
	server          *server.GrpcServer
	producer        *kafka_producer.KafkaProducer
	logger          *logger.Logger
	tracer          *oteltrace.TracerProvider
	metricsListener net.Listener
	shardManager    *shard_manager.ShardManager
}

// NewApp
func NewApp(ctx context.Context, cancel context.CancelFunc) (*App, error) {
	// Load environment
	_ = godotenv.Load()

	// Read flag
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = "config.yml"
	}

	// Read config
	cfg := config.NewConfig()
	if err := cfg.ReadConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Initialize logger
	var errorOutputPaths = []string{stdout}
	log := logger.NewLogger(ctx, cfg.Project.GetDebug(), errorOutputPaths, cfg.Project.GetName())

	// App info
	logger.Infow(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.GetName()),
		"version", cfg.Project.GetVersion(),
		"commitHash", cfg.Project.GetCommitHash(),
		"debug", cfg.Project.GetDebug(),
		"environment", cfg.Project.GetEnvironment(),
	)

	// Initialize tracer
	tr, err := tracer.InitTracer(ctx, cfg.Project.GetName(), cfg.Jaeger.GetURI())
	if err != nil {
		logger.Errorw(ctx, "Failed to initialize tracer", "error", err)
	}

	// Initialize DB connect
	pool, err := db.NewConnect(ctx, cfg.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize shard connections
	shardCount := len(cfg.Database.GetShards())
	if shardCount == 0 {
		return nil, fmt.Errorf("failed to connect to database: %w", internal_errors.ErrNoShardsAvailable)
	}

	shardPools := make([]*pgxpool.Pool, shardCount)
	for i, dsn := range cfg.Database.GetShards() {
		pool, err := db.NewConnect(ctx, dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shard %d: %w", i, err)
		}
		shardPools[i] = pool
	}

	// Initialize ShardManager
	shardFn := shard_manager.GetMurmur3ShardFn(shardCount)
	shardManager := shard_manager.NewShardManager(shardFn, shardPools, cfg.Database.GetShardBucketCount())

	// TxManager
	txManager := db.NewTransactionManager(pool)

	// Repository order/stocks/outbox
	repoOrder := repo_order.NewOrderRepository(shardManager)
	repoStocks := repo_stocks.NewStockRepository(pool)
	repoOutbox := repo_outbox.NewOutboxRepository(pool)

	// Kafka producer
	kafkaProd, err := kafka_producer.NewKafkaProducer(&cfg.Kafka)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Loms service
	lomsService := loms_service.NewService(repoOrder, repoStocks, repoOutbox, txManager, kafkaProd)

	// Loms app
	lomsApp := loms_app.NewService(lomsService)

	// GRPC server
	grpcServer := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, lomsApp)

	return &App{
		config:       cfg,
		ctx:          ctx,
		cancel:       cancel,
		pool:         pool,
		lomsService:  lomsService,
		lomsApp:      lomsApp,
		server:       grpcServer,
		producer:     kafkaProd,
		logger:       log,
		tracer:       tr,
		shardManager: shardManager,
	}, nil
}

// Run
func (a *App) Run() error {
	// Start metrics and profiling
	go a.startMetricsServer(a.config.Metrics.GetURI())

	// Run outbox
	go a.startOutboxProcessor()

	// Create channel to listen interrupt or terminate signals
	quitChan := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// GRPC server
	serverErrChan := make(chan error, errorChannelBufferSize)
	go func() {
		if err := a.server.Start(); err != nil {
			logger.Errorw(a.ctx, "Failed creating gRPC server", "error", err)
			serverErrChan <- err
		}
	}()

	// Wait
	select {
	case <-quitChan:
		logger.Infow(a.ctx, "Shutdown signal received")
	case err := <-serverErrChan:
		if err != nil {
			logger.Errorw(a.ctx, "Server error", "error", err)
			return err
		}
	}

	// Shutdown
	a.Shutdown()
	return nil
}

// startOutboxProcessor
func (a *App) startOutboxProcessor() {
	ticker := time.NewTicker(processOutboxInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.lomsService.ProcessOutbox(a.ctx); err != nil {
				logger.Errorw(a.ctx, "Error processing outbox", "error", err)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// Shutdown
func (a *App) Shutdown() {
	a.cancel()

	// Shutdown DB conn
	a.pool.Close()

	// Shutdown shard conn
	a.shardManager.CloseShards()

	if err := a.producer.Close(); err != nil {
		logger.Errorw(a.ctx, "Failed to close Kafka producer", "error", err)
	}

	// Shutdown metricsListener
	if a.metricsListener != nil {
		if err := a.metricsListener.Close(); err != nil {
			logger.Errorw(a.ctx, "Failed to close metrics listener", "error", err)
		}
	}

	// Shutdown tracer
	if err := a.tracer.Shutdown(a.ctx); err != nil {
		logger.Errorw(a.ctx, "Error shutting down tracer provider", "error", err)
	}

	// Shutdown logger
	if err := a.logger.Sync(); err != nil {
		logger.Errorw(a.ctx, "Failed to sync logger", "error", err)
	}

	logger.Infow(a.ctx, "Shutdown complete")
}

// startMetricsServer starts the metrics and profiling HTTP server.
func (a *App) startMetricsServer(uri string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))

	// Run
	listener, err := net.Listen("tcp4", uri)
	if err != nil {
		logger.Errorw(a.ctx, "Failed to start metrics server", "error", err)
		return
	}

	a.metricsListener = listener

	if err := http.Serve(listener, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorw(a.ctx, "Metrics server stopped", "error", err)
	}
}
