package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	internal_errors "route256/cart/internal/pkg/errors"
)

// Function for set response headers.
func setResponseHeaders(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
}

// Function for write JSON error.
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	setResponseHeaders(w, statusCode)
	_, errOut := fmt.Fprintf(w, "{\"message\":\"%s\"}", message)
	if errOut != nil {
		log.Printf("Response writing failed: %s", errOut.Error())
	}
}

// Function to determine the HTTP status of the code based on an error
func getStatusCodeFromError(err error) int {
	switch {
	case errors.Is(err, internal_errors.ErrBadRequest):
		return http.StatusBadRequest // 400
	case errors.Is(err, internal_errors.ErrUnauthorized):
		return http.StatusUnauthorized // 401
	case errors.Is(err, internal_errors.ErrForbidden):
		return http.StatusForbidden // 403
	case errors.Is(err, internal_errors.ErrNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, internal_errors.ErrInternalServerError):
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}
