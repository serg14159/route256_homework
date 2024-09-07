package server

import (
	"encoding/json"
	"io"
	"net/http"
	"route256/cart/internal/models"
	"strconv"
)

// Function handler for add product into cart.
func (s *Server) AddProduct(w http.ResponseWriter, r *http.Request) {
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	rawSKU := r.PathValue("sku_id")
	SKU, err := strconv.ParseInt(rawSKU, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

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

	err = s.cartService.AddProduct(r.Context(), UID, SKU, req.Count)
	if err != nil {
		writeJSONError(w, getStatusCodeFromError(err), err.Error())
		return
	}

	setResponseHeaders(w, http.StatusOK)
}
