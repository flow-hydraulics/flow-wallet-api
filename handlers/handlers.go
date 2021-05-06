// Package provides HTTP handlers for different services across the application.
package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
)

// handleError is a helper function for unified HTTP error handling.
func handleError(err error, log *log.Logger, rw http.ResponseWriter) {
	log.Printf("Error: %v\n", err)

	// Check if the error was an errors.RequestError
	reqErr, isReqErr := err.(*errors.RequestError)
	if isReqErr {
		// Send error message to client
		http.Error(rw, reqErr.Error(), reqErr.StatusCode)
		return
	}

	// Otherwise do not send data regarding the error
	http.Error(rw, "Error", http.StatusInternalServerError)
}

// handleJsonResponse is a helper function for unified JSON response handling.
func handleJsonResponse(rw http.ResponseWriter, status int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
}
