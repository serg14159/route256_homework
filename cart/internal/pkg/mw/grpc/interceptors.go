package mw

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

// GrpcUnaryClientInterceptor returns a new unary client interceptor that adds tracing.
func GrpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		// Create a span for the client call
		ctx, span := otel.Tracer("GrpcClient").Start(ctx, method)
		defer span.End()

		// Invoke the original method
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// GrpcStreamClientInterceptor returns a new stream client interceptor that adds tracing.
func GrpcStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {

		// Create a span for the client call
		ctx, span := otel.Tracer("GrpcClient").Start(ctx, method)
		defer span.End()

		// Invoke the original method
		return streamer(ctx, desc, cc, method, opts...)
	}
}
