package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	api "route256/loms/internal/app/loms"
	pb "route256/loms/pkg/api/loms/v1"

	mw "route256/loms/internal/pkg/mv"
)

const quitChannelBufferSize = 1

type ICfgPrj interface {
	GetDebug() bool
}

type ICfgGrpc interface {
	GetGrpcHost() string
	GetGrpcPort() int
	GetGrpcMaxConnectionIdle() int64
	GetGrpcTimeout() int64
	GetMaxConnectionAge() int64
}

type ICfgGateway interface {
	GetGatewayHost() string
	GetGatewayPort() int
	GetGatewayAllowedCORSOrigins() []string
}

type ICfgSwagger interface {
	GetSwaggerHost() string
	GetSwaggerPort() int
	GetGtAddr() string
	GetFilepath() string
	GetDist() string
}

type GrpcServer struct {
	cfgPrj         ICfgPrj
	cfgGrpc        ICfgGrpc
	cfgGateway     ICfgGateway
	cfgSwagger     ICfgSwagger
	lomsServiceApi *api.Service
	gatewayServer  *http.Server
	swaggerServer  *http.Server
	grpcServer     *grpc.Server
}

func NewGrpcServer(cfgPrj ICfgPrj, cfgGrpc ICfgGrpc, cfgGateway ICfgGateway, cfgSwagger ICfgSwagger, lomsServiceApi *api.Service) *GrpcServer {
	return &GrpcServer{
		cfgPrj:         cfgPrj,
		cfgGrpc:        cfgGrpc,
		cfgGateway:     cfgGateway,
		cfgSwagger:     cfgSwagger,
		lomsServiceApi: lomsServiceApi,
	}
}

// Start
func (s *GrpcServer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.startServers(ctx, cancel); err != nil {
		return err
	}

	s.awaitTermination(ctx)
	s.shutdownServers(ctx)
	return nil
}

// Start servers
func (s *GrpcServer) startServers(ctx context.Context, cancel context.CancelFunc) error {
	// Start Gateway server
	if err := s.startGatewayServer(ctx, cancel); err != nil {
		return err
	}

	// Start Swagger server
	if err := s.startSwaggerServer(ctx, cancel); err != nil {
		return err
	}

	// Start gRPC server
	return s.startGrpcServer(ctx)
}

func (s *GrpcServer) startGatewayServer(ctx context.Context, cancel context.CancelFunc) error {
	gatewayAddr := fmt.Sprintf("%s:%v", s.cfgGateway.GetGatewayHost(), s.cfgGateway.GetGatewayPort())
	grpcAddr := fmt.Sprintf("%s:%v", s.cfgGrpc.GetGrpcHost(), s.cfgGrpc.GetGrpcPort())
	s.gatewayServer = createGatewayServer(grpcAddr, gatewayAddr, s.cfgGateway.GetGatewayAllowedCORSOrigins())

	go func() {
		log.Printf("Gateway server is running on %s", gatewayAddr)
		if err := s.gatewayServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed running gateway server: %v", err)
			cancel()
		}
	}()
	return nil
}

func (s *GrpcServer) startSwaggerServer(ctx context.Context, cancel context.CancelFunc) error {
	swaggerAddr := fmt.Sprintf("%s:%v", s.cfgSwagger.GetSwaggerHost(), s.cfgSwagger.GetSwaggerPort())
	swaggerGtAddr := fmt.Sprintf("%s:%v", s.cfgSwagger.GetGtAddr(), s.cfgGateway.GetGatewayPort())
	var err error
	s.swaggerServer, err = createSwaggerServer(swaggerGtAddr, swaggerAddr, s.cfgSwagger.GetFilepath(), s.cfgSwagger.GetDist())
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Swagger server is running on %s", swaggerAddr)
		if err := s.swaggerServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed running swagger server: %v", err)
			cancel()
		}
	}()
	return nil
}

func (s *GrpcServer) startGrpcServer(ctx context.Context) error {
	grpcAddr := fmt.Sprintf("%s:%v", s.cfgGrpc.GetGrpcHost(), s.cfgGrpc.GetGrpcPort())
	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = s.createGrpcServer()
	go func() {
		log.Printf("GRPC Server is listening on: %s", grpcAddr)
		if err := s.grpcServer.Serve(l); err != nil {
			log.Printf("Failed running gRPC server: %v", err)
		}
	}()
	return nil
}

func (s *GrpcServer) createGrpcServer() *grpc.Server {
	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Duration(s.cfgGrpc.GetGrpcMaxConnectionIdle()) * time.Minute,
			Timeout:           time.Duration(s.cfgGrpc.GetGrpcTimeout()) * time.Second,
			MaxConnectionAge:  time.Duration(s.cfgGrpc.GetMaxConnectionAge()) * time.Minute,
			Time:              time.Duration(s.cfgGrpc.GetGrpcTimeout()) * time.Minute,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			mw.Logger,
			mw.Validate,
			grpcrecovery.UnaryServerInterceptor(),
		)),
	)

	pb.RegisterLomsServer(server, s.lomsServiceApi)

	if s.cfgPrj.GetDebug() {
		reflection.Register(server)
	}

	return server
}

func (s *GrpcServer) awaitTermination(ctx context.Context) {
	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		log.Printf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		log.Printf("ctx.Done: %v", done)
	}
}

func (s *GrpcServer) shutdownServers(ctx context.Context) {
	if s.gatewayServer != nil {
		if err := s.gatewayServer.Shutdown(ctx); err != nil {
			log.Printf("gatewayServer.Shutdown: %v", err)
		} else {
			log.Printf("gatewayServer shut down correctly")
		}
	}

	if s.swaggerServer != nil {
		if err := s.swaggerServer.Shutdown(ctx); err != nil {
			log.Printf("swaggerServer.Shutdown: %v", err)
		} else {
			log.Printf("swaggerServer shut down correctly")
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		log.Printf("grpcServer shut down correctly")
	}
}
