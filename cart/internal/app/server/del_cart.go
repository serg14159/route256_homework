package server

import (
	"log"
	"net/http"
	"strconv"
)

func (s *Server) DelCart(w http.ResponseWriter, r *http.Request) {
	log.Printf("DelCart")
	rawUID := r.PathValue("user_id")
	UID, err := strconv.ParseInt(rawUID, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("UID: %v", UID)

	if UID < 1 {
		writeJSONError(w, http.StatusBadRequest, "fail validation")
		return
	}

	err = s.cartService.DelCart(r.Context(), UID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("err: %v", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
