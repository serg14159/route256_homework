package server

import (
	"encoding/json"
	"net/http"
	"route256/cart/internal/models"
	"strconv"
)

// Function handler for get cart contents.
func (s *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if UID < 1 {
		writeJSONError(w, http.StatusBadRequest, "validation failed")
		return
	}

	items, totalPrice, err := s.cartService.GetCart(r.Context(), UID)
	if err != nil {
		writeJSONError(w, getStatusCodeFromError(err), err.Error())
		return
	}

	res := models.GetCartResponse{
		Items:      items,
		TotalPrice: totalPrice,
	}

	rawRes, err := json.Marshal(res)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	setResponseHeaders(w, http.StatusOK)
	w.Write(rawRes)
}
