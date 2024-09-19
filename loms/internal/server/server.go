package server

import (
	"context"
	"errors"
	"fmt"

	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	api "route256/loms/internal/app/loms"
	mw "route256/loms/internal/pkg/mv"
	pb "route256/loms/pkg/api/loms/v1"
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

// Function for start server.
func (s *GrpcServer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gatewayAddr := fmt.Sprintf("%s:%v", s.cfgGateway.GetGatewayHost(), s.cfgGateway.GetGatewayPort())
	swaggerAddr := fmt.Sprintf("%s:%v", s.cfgSwagger.GetSwaggerHost(), s.cfgSwagger.GetSwaggerPort())
	swaggerGtAddr := fmt.Sprintf("%s:%v", s.cfgSwagger.GetGtAddr(), s.cfgGateway.GetGatewayPort())
	grpcAddr := fmt.Sprintf("%s:%v", s.cfgGrpc.GetGrpcHost(), s.cfgGrpc.GetGrpcPort())

	gatewayServer := createGatewayServer(grpcAddr, gatewayAddr, s.cfgGateway.GetGatewayAllowedCORSOrigins())
	swaggerServer, err := createSwaggerServer(swaggerGtAddr, swaggerAddr, s.cfgSwagger.GetFilepath(), s.cfgSwagger.GetDist())
	if err != nil {
		return err
	}

	go func() {
		log.Printf("Gateway server is running on %s", gatewayAddr)
		if err := gatewayServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed running gateway server: %v", err)
			cancel()
		}
	}()

	go func() {
		log.Printf("Swagger server is running on %s", swaggerAddr)
		if err := swaggerServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed running swagger server: %v", err)
			cancel()
		}
	}()

	l, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer l.Close()

	grpcServer := grpc.NewServer(
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

	pb.RegisterLomsServer(grpcServer, s.lomsServiceApi)

	go func() {
		log.Printf("GRPC Server is listening on: %s", grpcAddr)
		if err := grpcServer.Serve(l); err != nil {
			log.Printf("Failed running gRPC server: %v", err)
		}
	}()

	if s.cfgPrj.GetDebug() {
		reflection.Register(grpcServer)
	}

	quit := make(chan os.Signal, quitChannelBufferSize)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		log.Printf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		log.Printf("ctx.Done: %v", done)
	}

	if err := gatewayServer.Shutdown(ctx); err != nil {
		log.Printf("gatewayServer.Shutdown: %v", err)
	} else {
		log.Printf("gatewayServer shut down correctly: %v", err)
	}

	grpcServer.GracefulStop()
	log.Printf("grpcServer shut down correctly")

	return nil
}
