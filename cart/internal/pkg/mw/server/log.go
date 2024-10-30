package mw

import (
	"context"
	"net/http"
	"time"

	"route256/cart/internal/pkg/metrics"
	"route256/utils/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Middleware struct that wraps the HTTP handler.
type Middleware struct {
	handler http.Handler
}

// New creates new Middleware instance.
func New(handler http.Handler) http.Handler {
	return &Middleware{handler: handler}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ServeHTTP implements the http.Handler interface.
func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Wrap ResponseWriter to capture status code
	rw := &responseWriter{w, http.StatusOK}

	// Updated context
	var ctx context.Context

	// Wrap the handler to capture the context
	handlerWithContextCapture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
		m.handler.ServeHTTP(w, r)
	})

	// Wrap handler with OpenTelemetry middleware
	otelHandler := otelhttp.NewHandler(handlerWithContextCapture, "HTTP Server")
	otelHandler.ServeHTTP(rw, r)

	duration := time.Since(start)
	handler := r.URL.Path
	statusCode := rw.statusCode

	if ctx == nil {
		ctx = r.Context()
	}

	// Increment metrics
	metrics.IncRequestCounterWithStatus(handler, statusCode)
	metrics.ObserveHandlerDuration(handler, duration)

	// Log request
	logger.Infow(ctx, "Handled request",
		"method", r.Method,
		"url", handler,
		"status", statusCode,
		"duration", duration.Seconds(),
	)
}
