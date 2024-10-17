package server

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mw "route256/loms/internal/pkg/mw"
	pb "route256/loms/pkg/api/loms/v1"
)

// createGatewayServer create gateway server.
func createGatewayServer(grpcAddr, gatewayAddr string, allowedOrigins []string) *http.Server {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to dial server: %v", err)
	}

	mux := runtime.NewServeMux()
	if err := pb.RegisterLomsHandler(context.Background(), mux, conn); err != nil {
		log.Printf("Failed registration handler: %v", err)
	}

	middlewareChain := сhainHTTPMiddleware(
		mw.WithHTTPLoggingMiddleware,
		mw.Cors(allowedOrigins),
	)

	gatewayServer := &http.Server{
		Addr:    gatewayAddr,
		Handler: middlewareChain(mux),
	}

	return gatewayServer
}

// сhainHTTPMiddleware combines several HTTP middleware into one chain.
func сhainHTTPMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(finalHandler http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			finalHandler = middlewares[i](finalHandler)
		}
		return finalHandler
	}
}
