package mw

import (
	"net/http"
)

// Function CorsMiddleware create CORS mv.
func Cors(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			providedOrigin := r.Header.Get("Origin")
			if isAllowedOrigin(providedOrigin, allowedOrigins) {
				setCORSHeaders(w, providedOrigin)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Function isAllowedOrigin checks if provided origin is in allowed origins list.
func isAllowedOrigin(providedOrigin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if providedOrigin == allowedOrigin {
			return true
		}
	}
	return false
}

// Function setCORSHeaders sets the necessary CORS headers.
func setCORSHeaders(w http.ResponseWriter, origin string) {
	headers := map[string]string{
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE",
		"Access-Control-Allow-Headers":     "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
