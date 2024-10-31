package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"route256/cart/internal/app/server"
	"route256/cart/internal/clients/product_service"
	"route256/cart/internal/config"

	"route256/utils/logger"

	loms_service "route256/cart/internal/clients/loms"
	"route256/cart/internal/pkg/cacher"
	grpc_mw "route256/cart/internal/pkg/mw/grpc"
	cart_repository "route256/cart/internal/repository/cart"
	cart_service "route256/cart/internal/service/cart"
	"route256/utils/tracer"

	oteltrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownTimeout = 5 * time.Second
const stdout = "stdout"

type App struct {
	config          *config.Config
	logger          *logger.Logger
	tracer          *oteltrace.TracerProvider
	server          *server.Server
	cartService     *cart_service.CartService
	metricsListener net.Listener
	lomsClient      *loms_service.LomsClient
	productClient   *product_service.Client
	connGrpc        *grpc.ClientConn
}

// NewApp
func NewApp(ctx context.Context) (*App, error) {
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

	// Init logger
	var errorOutputPaths = []string{stdout}
	log := logger.NewLogger(ctx, cfg.Project.GetDebug(), errorOutputPaths, cfg.Project.GetName())

	// App info
	logger.Infow(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.GetName()),
		"version", cfg.Project.GetVersion(),
		"commitHash", cfg.Project.GetCommitHash(),
		"debug", cfg.Project.GetDebug(),
		"environment", cfg.Project.GetEnvironment(),
	)

	// Init tracer
	tr, err := tracer.InitTracer(ctx, cfg.Project.GetName(), cfg.Jaeger.GetURI())
	if err != nil {
		logger.Errorw(ctx, "Failed to initialize tracer", "error", err)
	}

	// Init repository
	cartRepository := cart_repository.NewCartRepository()

	// Product service client
	productService := product_service.NewClient(&cfg.ProductService)

	// Cacher
	cacher := cacher.NewLRUCache(cfg.Cache.Capacity)

	// Product service client with cache
	productServiceWithCache := product_service.NewClientWithCache(productService, cacher)

	// Loms service client
	lomsAddr := fmt.Sprintf("%s:%s", cfg.LomsService.GetHost(), cfg.LomsService.GetPort())
	connGrpc, err := grpc.NewClient(lomsAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_mw.GrpcUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpc_mw.GrpcStreamClientInterceptor()),
	)
	if err != nil {
		logger.Errorw(ctx, "Did not connect", "error", err)
	}

	loms := loms_service.NewLomsClient(connGrpc)

	// Init service
	cartService := cart_service.NewService(cartRepository, productServiceWithCache, loms)

	// Init server
	srv := server.NewServer(&cfg.Server, cartService)

	return &App{
		config:        cfg,
		logger:        log,
		tracer:        tr,
		server:        srv,
		cartService:   cartService,
		lomsClient:    loms,
		productClient: productService,
		connGrpc:      connGrpc,
	}, nil
}

func (a *App) Run() error {
	// Start metrics and profiling
	go a.startMetricsServer(a.config.Metrics.GetURI())

	// Run server
	if err := a.server.Run(); err != nil {
		logger.Errorw(context.Background(), "Failed to start server", "error", err)
		return err
	}

	return nil
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Errorw(ctx, "Failed to shutdown server", "error", err)
		return err
	}

	// Shutdown metricsListener
	if a.metricsListener != nil {
		if err := a.metricsListener.Close(); err != nil {
			logger.Errorw(ctx, "Failed to close metrics listener", "error", err)
		}
	}

	// Shutdown tracer
	if err := a.tracer.Shutdown(ctx); err != nil {
		logger.Errorw(ctx, "Error shutting down tracer provider", "error", err)
	}

	// Shutdown logger
	if err := a.logger.Sync(); err != nil {
		logger.Errorw(ctx, "Failed to sync logger", "error", err)
	}

	a.connGrpc.Close()

	return nil
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
		logger.Errorw(context.Background(), "Failed to start metrics server", "error", err)
		return
	}

	a.metricsListener = listener

	if err := http.Serve(listener, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorw(context.Background(), "Metrics server stopped", "error", err)
	}
}
