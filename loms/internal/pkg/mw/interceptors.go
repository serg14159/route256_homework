package mw

import (
	"context"
	"route256/loms/internal/pkg/metrics"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerMetricsInterceptor collects metrics for unary gRPC requests.
func UnaryServerMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		duration := time.Since(start)

		statusCode := status.Code(err).String()

		metrics.IncRequestCounterWithStatus(info.FullMethod, statusCode)
		metrics.ObserveHandlerDuration(info.FullMethod, duration)

		return resp, err
	}
}

// StreamServerMetricsInterceptor collects metrics for stream gRPC requests.
func StreamServerMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()
		err := handler(srv, ss)
		duration := time.Since(startTime)

		statusCode := status.Code(err).String()

		// Increment request counter with status code
		metrics.IncRequestCounterWithStatus(info.FullMethod, statusCode)

		// Observe handler duration
		metrics.ObserveHandlerDuration(info.FullMethod, duration)

		return err
	}
}
