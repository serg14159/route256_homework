package server

import (
	"log"
	"net/http"
	"strconv"
)

// Function handler for delete product from cart.
func (s *Server) DelProduct(w http.ResponseWriter, r *http.Request) {
	log.Printf("DelProduct")
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("UID: %v", UID)

	rawSKU := r.PathValue("sku_id")
	SKU, err := strconv.ParseInt(rawSKU, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("SKU: %v", SKU)

	if UID < 1 || SKU < 1 {
		writeJSONError(w, http.StatusBadRequest, "fail validation")
		return
	}

	err = s.cartService.DelProduct(r.Context(), UID, SKU)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
