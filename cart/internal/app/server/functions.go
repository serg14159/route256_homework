package server

import (
	"fmt"
	"log"
	"net/http"
)

func setResponseHeaders(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	setResponseHeaders(w, statusCode)
	_, errOut := fmt.Fprintf(w, "{\"message\":\"%s\"}", message)
	if errOut != nil {
		log.Printf("Response writing failed: %s", errOut.Error())
	}
}
