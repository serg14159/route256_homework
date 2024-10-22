package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// requestCounter tracks the total number of requests handled by each handler.
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "handler_request_total_counter",
			Help:      "Total number of requests by handler",
		},
		[]string{"handler"},
	)

	// handlerHistogram tracks the duration of request handling for each handler.
	handlerHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "handler_duration_histogram",
			Help:      "Total duration of handler processing by request",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"handler"},
	)

	// dbOperationsCounter tracks the total number of database operations.
	dbOperationsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "db_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"operation"},
	)

	// dbLatencyHistogram tracks the latency of each database operation.
	dbLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "db_operations_latency_seconds",
			Help:      "Latency of database operations",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// inMemoryItemsGauge tracks the total number of items in the in-memory repository.
	inMemoryItemsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "app",
			Name:      "in_memory_repository_items_total",
			Help:      "Total number of items in in-memory repository",
		},
	)
)

// IncRequestCounter increments the request counter for a specific handler.
func IncRequestCounter(handler string) {
	requestCounter.WithLabelValues(handler).Inc()
}

// ObserveHandlerDuration observes and records the duration of a handler is execution.
func ObserveHandlerDuration(handler string, duration time.Duration) {
	handlerHistogram.WithLabelValues(handler).Observe(duration.Seconds())
}

// IncDBOperation increments the counter for a specific database operation.
func IncDBOperation(operation string) {
	dbOperationsCounter.WithLabelValues(operation).Inc()
}

// ObserveDBLatency observes and records the latency of a database operation.
func ObserveDBLatency(operation string, duration time.Duration) {
	dbLatencyHistogram.WithLabelValues(operation).Observe(duration.Seconds())
}

// SetInMemoryItemsTotal sets the current number of items in the in-memory repository.
func SetInMemoryItemsTotal(count int) {
	inMemoryItemsGauge.Set(float64(count))
}
