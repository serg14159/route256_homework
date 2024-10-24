package server

import (
	"net/http"
	"strconv"
)

// DelProduct handler for delete product from cart.
func (s *Server) DelProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	err = s.cartService.DelProduct(ctx, UID, SKU)
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	setResponseHeaders(w, http.StatusNoContent)
}
