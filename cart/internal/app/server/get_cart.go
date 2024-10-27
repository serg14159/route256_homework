package server

import (
	"encoding/json"
	"net/http"
	"route256/cart/internal/models"
	"strconv"

	"go.opentelemetry.io/otel"
)

// GetCart handler for get cart contents.
func (s *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	// Context
	ctx := r.Context()

	// Tracer
	ctx, span := otel.Tracer("CartHandlers").Start(ctx, "GetCart")
	defer span.End()

	// Get and check req
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	if UID < 1 {
		writeJSONError(ctx, w, http.StatusBadRequest, "validation failed")
		return
	}

	// Call service
	items, totalPrice, err := s.cartService.GetCart(ctx, UID)
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	res := models.GetCartResponse{
		Items:      items,
		TotalPrice: totalPrice,
	}

	rawRes, err := json.Marshal(res)
	if err != nil {
		writeJSONError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	setResponseHeaders(w, http.StatusOK)
	w.Write(rawRes)
}
