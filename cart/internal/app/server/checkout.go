package server

import (
	"encoding/json"
	"net/http"
	"route256/cart/internal/models"
	"strconv"
)

// Checkout handler for processes order request.
func (s *Server) Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	orderID, err := s.cartService.Checkout(ctx, models.UID(UID))
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	res := models.CheckoutResponse{
		OrderID: orderID,
	}

	rawRes, err := json.Marshal(res)
	if err != nil {
		writeJSONError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	setResponseHeaders(w, http.StatusOK)
	w.Write(rawRes)
}
