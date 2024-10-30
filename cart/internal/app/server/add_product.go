package server

import (
	"encoding/json"
	"io"
	"net/http"
	"route256/cart/internal/models"
	"strconv"

	"go.opentelemetry.io/otel"
)

// AddProduct handler for add product into cart.
func (s *Server) AddProduct(w http.ResponseWriter, r *http.Request) {
	// Context
	ctx := r.Context()

	// Tracer
	ctx, span := otel.Tracer("CartHandlers").Start(ctx, "AddProduct")
	defer span.End()

	// Get and check req
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	rawSKU := r.PathValue("sku_id")
	SKU, err := strconv.ParseInt(rawSKU, 10, 64)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	if UID < 1 || SKU < 1 {
		writeJSONError(ctx, w, http.StatusBadRequest, "validation failed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.AddProductRequest

	err = json.Unmarshal(body, &req)
	if err != nil {
		writeJSONError(ctx, w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Call service
	err = s.cartService.AddProduct(ctx, UID, SKU, req.Count)
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	setResponseHeaders(w, http.StatusOK)
}
