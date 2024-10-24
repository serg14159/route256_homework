package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "handler_request_total_counter",
			Help:      "Total number of requests by handler and status code",
		},
		[]string{"handler", "status"},
	)

	handlerHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "handler_duration_histogram",
			Help:      "Total duration of handler processing by request",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"handler"},
	)

	externalRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "external_request_total_counter",
			Help:      "Total number of external requests",
		},
		[]string{"url", "status"},
	)

	externalRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "external_request_duration_seconds",
			Help:      "Duration of external requests",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"url"},
	)

	dbOperationsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "db_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"operation"},
	)

	dbLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "db_operations_latency_seconds",
			Help:      "Latency of database operations",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation", "status"},
	)
)

// IncRequestCounterWithStatus increments the request counter for a handler with status code.
func IncRequestCounterWithStatus(handler string, statusCode int) {
	status := strconv.Itoa(statusCode)
	requestCounter.WithLabelValues(handler, status).Inc()
}

// ObserveHandlerDuration records the duration of a handler execution.
func ObserveHandlerDuration(handler string, duration time.Duration) {
	handlerHistogram.WithLabelValues(handler).Observe(duration.Seconds())
}

// IncExternalRequestCounter increments the external request counter.
func IncExternalRequestCounter(url string, status string) {
	externalRequestCounter.WithLabelValues(url, status).Inc()
}

// ObserveExternalRequestDuration records the duration of an external request.
func ObserveExternalRequestDuration(url string, duration time.Duration) {
	externalRequestDuration.WithLabelValues(url).Observe(duration.Seconds())
}

// IncDBOperation increments the counter for a database operation.
func IncDBOperation(operation string) {
	dbOperationsCounter.WithLabelValues(operation).Inc()
}

// ObserveDBLatency records the latency of a database operation.
func ObserveDBLatency(operation string, duration time.Duration, status string) {
	dbLatencyHistogram.WithLabelValues(operation, status).Observe(duration.Seconds())
}
