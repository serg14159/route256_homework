package server

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "route256/loms/pkg/api/loms/v1"
)

func createGatewayServer(grpcAddr, gatewayAddr string, allowedOrigins []string) *http.Server {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		context.Background(),
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("Failed to dial server: %v", err)
	}

	mux := runtime.NewServeMux()
	if err := pb.RegisterLomsHandler(context.Background(), mux, conn); err != nil {
		log.Printf("Failed registration handler: %v", err)
	}

	gatewayServer := &http.Server{
		Addr:    gatewayAddr,
		Handler: cors(mux, allowedOrigins),
	}

	return gatewayServer
}

func cors(h http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedOrigin := r.Header.Get("Origin")
		matches := false
		for _, allowedOrigin := range allowedOrigins {
			if providedOrigin == allowedOrigin {
				matches = true
				break
			}
		}

		if matches {
			w.Header().Set("Access-Control-Allow-Origin", providedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
		}
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
