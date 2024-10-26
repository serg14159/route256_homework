package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	loms "route256/loms/internal/app/loms"
	"route256/loms/internal/config"
	db "route256/loms/internal/pkg/db"
	kafkaProducer "route256/loms/internal/pkg/kafka"
	"route256/loms/internal/pkg/logger"
	loggerPkg "route256/loms/internal/pkg/logger"
	"route256/loms/internal/pkg/tracer"
	repo_order "route256/loms/internal/repository/orders"
	repo_outbox "route256/loms/internal/repository/outbox"
	repo_stocks "route256/loms/internal/repository/stocks"
	"route256/loms/internal/server"
	loms_usecase "route256/loms/internal/service/loms"

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
	usecase         *loms_usecase.LomsService
	controller      *loms.Service
	server          *server.GrpcServer
	producer        *kafkaProducer.KafkaProducer
	logger          *logger.Logger
	tracer          *oteltrace.TracerProvider
	metricsListener net.Listener
}

// NewApp
func NewApp() (*App, error) {
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
		return nil, fmt.Errorf("failed to init configuration: %w", err)
	}

	// Context
	ctx, cancel := context.WithCancel(context.Background())

	// Init logger
	var errorOutputPaths = []string{stdout}
	logger := loggerPkg.NewLogger(ctx, cfg.Project.GetDebug(), errorOutputPaths, cfg.Project.GetName())

	// App info
	loggerPkg.Infow(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.GetName()),
		"version", cfg.Project.GetVersion(),
		"commitHash", cfg.Project.GetCommitHash(),
		"debug", cfg.Project.GetDebug(),
		"environment", cfg.Project.GetEnvironment(),
	)

	// Init tracer
	tp, err := tracer.InitTracer(ctx, cfg.Project.GetName(), cfg.Jaeger.GetURI())
	if err != nil {
		loggerPkg.Errorw(ctx, "Failed to initialize tracer", "error", err)
	}

	// DB connect
	pool, err := db.NewConnect(ctx, &cfg.Database)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// TxManager
	txManager := db.NewTransactionManager(pool)

	// Repository order/stocks/outbox
	repoOrder := repo_order.NewOrderRepository(pool)
	repoStocks := repo_stocks.NewStockRepository(pool)
	repoOutbox := repo_outbox.NewOutboxRepository(pool)

	// Kafka producer
	kafkaProd, err := kafkaProducer.NewKafkaProducer(&cfg.Kafka)
	if err != nil {
		pool.Close()
		cancel()
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Loms usecase
	lomsUsecaseService := loms_usecase.NewService(repoOrder, repoStocks, repoOutbox, txManager, kafkaProd)

	// Loms controller
	controller := loms.NewService(lomsUsecaseService)

	// GRPC server
	grpcServer := server.NewGrpcServer(&cfg.Project, &cfg.Grpc, &cfg.Gateway, &cfg.Swagger, controller)

	return &App{
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
		pool:       pool,
		usecase:    lomsUsecaseService,
		controller: controller,
		server:     grpcServer,
		producer:   kafkaProd,
		logger:     logger,
		tracer:     tp,
	}, nil
}

// Run
func (a *App) Run() error {
	// Start metrics and profiling
	go a.startMetricsServer()

	// Run outbox
	go a.startOutboxProcessor()

	// Create channel to listen interrupt or terminate signals
	quitChan := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	// GRPC server
	serverErrChan := make(chan error, errorChannelBufferSize)
	go func() {
		if err := a.server.Start(); err != nil {
			log.Printf("Failed creating gRPC server, err:%s", err)
			serverErrChan <- err
		}
	}()

	// Wait
	select {
	case <-quitChan:
		log.Println("Shutdown signal received")
	case err := <-serverErrChan:
		if err != nil {
			log.Printf("Server error: %v", err)
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
			if err := a.usecase.ProcessOutbox(a.ctx); err != nil {
				log.Printf("Error processing outbox: %v", err)
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

	if err := a.producer.Close(); err != nil {
		log.Printf("Failed to close Kafka producer: %v", err)
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

	log.Printf("Shutdown complete")
}

// startMetricsServer starts the metrics and profiling HTTP server.
func (a *App) startMetricsServer() {
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
	listener, err := net.Listen("tcp4", "0.0.0.0:2113")
	if err != nil {
		logger.Errorw(context.Background(), "Failed to start metrics server", "error", err)
		return
	}

	a.metricsListener = listener

	if err := http.Serve(listener, mux); err != nil && err != http.ErrServerClosed {
		logger.Errorw(context.Background(), "Metrics server stopped", "error", err)
	}
}
