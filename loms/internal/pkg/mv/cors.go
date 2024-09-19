package mw

import (
	"net/http"
)

// Function CorsMiddleware create CORS mv.
func Cors(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedOrigin := r.Header.Get("Origin")
			matches := false
			for _, allowedOrigin := range allowedOrigins {
				if providedOrigin == allowedOrigin {
					matches = true
					break
				}
			}

			if matches {
				w.Header().Set("Access-Control-Allow-Origin", providedOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
