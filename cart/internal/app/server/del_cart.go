package server

import (
	"net/http"
	"strconv"
)

// DelCart handler for delete user cart.
func (s *Server) DelCart(w http.ResponseWriter, r *http.Request) {
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

	err = s.cartService.DelCart(ctx, UID)
	if err != nil {
		writeJSONError(ctx, w, getStatusCodeFromError(err), err.Error())
		return
	}

	setResponseHeaders(w, http.StatusNoContent)
}
