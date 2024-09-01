package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"route256/cart/internal/models"
	"strconv"
)

// Function handler for add product into cart.
func (s *Server) AddProduct(w http.ResponseWriter, r *http.Request) {
	log.Printf("AddProduct")
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
		writeJSONError(w, http.StatusBadRequest, "validation failed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.AddProductRequest

	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("req.Count: %v", req.Count)

	err = s.cartService.AddProduct(r.Context(), UID, SKU, req.Count)
	if err != nil {
		writeJSONError(w, getStatusCodeFromError(err), err.Error())
		return
	}

	log.Printf("req.Count: %v", req.Count)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
