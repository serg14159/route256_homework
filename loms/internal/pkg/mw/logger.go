package mw

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"route256/loms/internal/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Logger create request/response logger mv.
func Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	raw, _ := protojson.Marshal((req).(proto.Message))
	logger.Infow(ctx, fmt.Sprintf("request: method: %v, req: %v\n", info.FullMethod, string(raw)))

	if resp, err = handler(ctx, req); err != nil {
		logger.Infow(ctx, fmt.Sprintf("response: method: %v, err: %v\n", info.FullMethod, err))
		return
	}

	rawResp, _ := protojson.Marshal((resp).(proto.Message))
	logger.Infow(ctx, fmt.Sprintf("response: method: %v, resp: %v\n", info.FullMethod, string(rawResp)))

	return
}

// WithHTTPLoggingMiddleware logging http request.
func WithHTTPLoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
