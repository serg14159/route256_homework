package server

import (
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
)

// DelCart handler for delete user cart.
func (s *Server) DelCart(w http.ResponseWriter, r *http.Request) {
	// Context
	ctx := r.Context()

	// Tracer
	ctx, span := otel.Tracer("CartHandlers").Start(ctx, "DelCart")
	defer span.End()

	// Get and check req
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	if UID < 1 {
		writeJSONError(ctx, w, http.StatusBadRequest, "validation fail")
		return
	}

	// Call service
	err = s.cartService.DelCart(ctx, UID)
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	setResponseHeaders(w, http.StatusNoContent)
}
