package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "request_total",
			Help:      "Total number of requests",
		},
		[]string{"method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "request_duration_seconds",
			Help:      "Duration of requests in seconds",
		},
		[]string{"method"},
	)

	DBQueryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "db_query_total",
			Help:      "Total number of database queries",
		},
		[]string{"query_type"},
	)

	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "db_query_duration_seconds",
			Help:      "Duration of database queries in seconds",
		},
		[]string{"query_type"},
	)
)

// IncRequestCounterWithStatus increments the request counter for a handler with status code.
func IncRequestCounterWithStatus(method string, status string) {
	RequestCounter.WithLabelValues(method, status).Inc()
}

// ObserveHandlerDuration records the duration of a handler execution.
func ObserveHandlerDuration(method string, duration time.Duration) {
	RequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// IncDBQueryCounter increments DB request counter.
func IncDBQueryCounter(queryType string) {
	DBQueryCounter.WithLabelValues(queryType).Inc()
}

// ObserveDBQueryDuration records the duration of DB request.
func ObserveDBQueryDuration(queryType string, duration time.Duration) {
	DBQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}
