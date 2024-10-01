package server_middleware

import (
	"log"
	"net/http"
	"time"
)

type LogMux struct {
	h http.Handler
}

func NewLogMux(h http.Handler) http.Handler {
	return &LogMux{h: h}
}

// ServeHTTP middleware for log http request.
func (m *LogMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	m.h.ServeHTTP(w, r)
	log.Printf("Request: %s %s took %v", r.Method, r.URL.Path, time.Since(start))
}
