package server

import (
	"net/http"
	"strconv"
)

// Function handler for delete user cart.
func (s *Server) DelCart(w http.ResponseWriter, r *http.Request) {
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if UID < 1 {
		writeJSONError(w, http.StatusBadRequest, "validation fail")
		return
	}

	err = s.cartService.DelCart(r.Context(), UID)
	if err != nil {
		writeJSONError(w, getStatusCodeFromError(err), err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
