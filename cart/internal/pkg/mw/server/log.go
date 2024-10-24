package middleware

import (
	"net/http"
	"time"

	"route256/cart/internal/pkg/logger"
	"route256/cart/internal/pkg/metrics"

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

	// Add logger to context
	ctx := logger.ToContext(r.Context(), logger.FromContext(r.Context()))
	r = r.WithContext(ctx)

	// Wrap ResponseWriter to capture status code
	rw := &responseWriter{w, http.StatusOK}

	// Wrap handler with OpenTelemetry middleware
	otelHandler := otelhttp.NewHandler(m.handler, "HTTP Server")
	otelHandler.ServeHTTP(rw, r)

	duration := time.Since(start)
	handler := r.URL.Path
	statusCode := rw.statusCode

	// Increment metrics
	metrics.IncRequestCounterWithStatus(handler, statusCode)
	metrics.ObserveHandlerDuration(handler, duration)

	// Log request
	logger.Infow(r.Context(), "Handled request",
		"method", r.Method,
		"url", handler,
		"status", statusCode,
		"duration", duration.Seconds(),
	)
}
