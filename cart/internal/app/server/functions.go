package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	internal_errors "route256/cart/internal/pkg/errors"

	"github.com/go-playground/validator"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// setResponseHeaders function for set response headers.
func setResponseHeaders(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
}

// writeJSONError function for write JSON error.
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	setResponseHeaders(w, statusCode)
	_, errOut := fmt.Fprintf(w, "{\"message\":\"%s\"}", message)
	if errOut != nil {
		log.Printf("Response writing failed: %s", errOut.Error())
	}
}

// getStatusCodeFromError function to determine the HTTP status of the code based on an error
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
	case errors.Is(err, internal_errors.ErrPreconditionFailed):
		return http.StatusPreconditionFailed // 412
	case errors.Is(err, internal_errors.ErrInternalServerError):
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}
