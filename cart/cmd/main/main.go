package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"route256/cart/internal/app/server"
	"route256/cart/internal/clients/product_service"
	"route256/cart/internal/config"
	"route256/cart/internal/pkg/logger"
	"route256/cart/internal/pkg/tracer"
	repository "route256/cart/internal/repository/cart"
	service "route256/cart/internal/service/cart"
	"time"

	"log"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	loms_service "route256/cart/internal/clients/loms"
)

const quitChannelBufferSize = 1
const shutdownTimeout = 5 * time.Second
const stdout = "stdout"

func main() {
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

	// Init context
	ctx := context.Background()

	// Initialize logger
	var errorOutputPaths = []string{stdout}
	logger := logger.NewLogger(ctx, cfg.Project.Debug, errorOutputPaths)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %s", err)
		}
	}()

	// App info
	logger.Infow(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.GetName()),
		"version", cfg.Project.GetVersion(),
		"commitHash", cfg.Project.GetCommitHash(),
		"debug", cfg.Project.GetDebug(),
		"environment", cfg.Project.GetEnvironment(),
	)

	// Add logger to context
	ctx = logger.ToContext(ctx, logger)

	// Initialize tracer
	tp, err := tracer.InitTracer(ctx, cfg.Project.GetName(), cfg.Jaeger.GetURI())
	if err != nil {
		logger.Errorw(ctx, "Failed to initialize tracer", "error", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Errorw(ctx, "Error shutting down tracer provider", "error", err)
		}
	}()

	// Repository
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
		logger.Errorw(ctx, "Did not connect", "error", err)
	}
	defer connGrpc.Close()

	loms := loms_service.NewLomsClient(connGrpc)

	// Service
	cartService := service.NewService(cartRepository, productService, loms)

	// Server
	s := server.NewServer(&cfg.Server, cartService)

	// Start metrics and profiling
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// pprof
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Errorw(ctx, "Failed to start metrics server", "error", err)
		}
	}()

	// Run server
	err = s.Run()
	if err != nil {
		logger.Errorw(ctx, "Failed to start server", "error", err)
	}

	// Wait os interrupt
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Infow(ctx, "Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Errorw(ctx, "Failed server shutdown", "error", err)
	}
	logger.Infow(ctx, "Server exiting")
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
