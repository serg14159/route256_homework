package mw

import (
	"context"
	"route256/loms/internal/pkg/metrics"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

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
