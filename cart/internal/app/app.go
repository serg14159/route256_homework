package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"route256/cart/internal/app/server"
	"route256/cart/internal/clients/product_service"
	"route256/cart/internal/config"
	"route256/cart/internal/pkg/logger"

	loms_service "route256/cart/internal/clients/loms"
	loggerPkg "route256/cart/internal/pkg/logger"
	"route256/cart/internal/pkg/tracer"
	repository "route256/cart/internal/repository/cart"
	service "route256/cart/internal/service/cart"

	oteltrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
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
	cartService     *service.CartService
	metricsListener net.Listener
	lomsClient      *loms_service.LomsClient
	productClient   *product_service.Client
	connGrpc        *grpc.ClientConn
}

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
		log.Printf("Failed init configuration, err:%s", err)
	}

	// Init logger
	var errorOutputPaths = []string{stdout}
	logger := loggerPkg.NewLogger(ctx, cfg.Project.Debug, errorOutputPaths)

	// App info
	loggerPkg.Infow(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.GetName()),
		"version", cfg.Project.GetVersion(),
		"commitHash", cfg.Project.GetCommitHash(),
		"debug", cfg.Project.GetDebug(),
		"environment", cfg.Project.GetEnvironment(),
	)

	// Add logger to context
	ctx = loggerPkg.ToContext(ctx, logger)

	// Init tracer
	tp, err := tracer.InitTracer(ctx, cfg.Project.GetName(), cfg.Jaeger.GetURI())
	if err != nil {
		loggerPkg.Errorw(ctx, "Failed to initialize tracer", "error", err)
	}

	// Init repository
	cartRepository := repository.NewCartRepository()

	// Product service client
	productService := product_service.NewClient(&cfg.ProductService)

	// Loms service client
	lomsAddr := fmt.Sprintf("%s:%s", cfg.LomsService.Host, cfg.LomsService.Port)
	connGrpc, err := grpc.NewClient(lomsAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpcStreamClientInterceptor()),
	)
	if err != nil {
		loggerPkg.Errorw(ctx, "Did not connect", "error", err)
	}

	loms := loms_service.NewLomsClient(connGrpc)

	// Init service
	cartService := service.NewService(cartRepository, productService, loms)

	// Init server
	srv := server.NewServer(&cfg.Server, cartService)

	app := &App{
		config:        cfg,
		logger:        logger,
		tracer:        tp,
		server:        srv,
		cartService:   cartService,
		lomsClient:    loms,
		productClient: productService,
		connGrpc:      connGrpc,
	}

	return app, nil
}

func (a *App) Run() error {
	// Start metrics and profiling
	go a.startMetricsServer()

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
	listener, err := net.Listen("tcp4", "0.0.0.0:2112")
	if err != nil {
		logger.Errorw(context.Background(), "Failed to start metrics server", "error", err)
		return
	}

	a.metricsListener = listener

	if err := http.Serve(listener, mux); err != nil && err != http.ErrServerClosed {
		logger.Errorw(context.Background(), "Metrics server stopped", "error", err)
	}
}

// grpcUnaryClientInterceptor returns a new unary client interceptor that adds tracing.
func grpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		// Create a span for the client call
		tracer := otel.Tracer("grpc-client")
		ctx, span := tracer.Start(ctx, method)
		defer span.End()

		// Invoke the original method
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// grpcStreamClientInterceptor returns a new stream client interceptor that adds tracing.
func grpcStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {

		// Create a span for the client call
		tracer := otel.Tracer("grpc-client")
		ctx, span := tracer.Start(ctx, method)
		defer span.End()

		// Invoke the original method
		return streamer(ctx, desc, cc, method, opts...)
	}
}
