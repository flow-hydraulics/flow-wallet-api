// Package handlers provides HTTP handlers for different services across the application.
package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
)

const SYNC_HEADER = "Use-Sync"

// handleError is a helper function for unified HTTP error handling.
func handleError(err error, logger *log.Logger, rw http.ResponseWriter) {
	logger.Printf("Error: %v\n", err)

	// Check if the error was an errors.RequestError
	reqErr, isReqErr := err.(*errors.RequestError)
	if isReqErr {
		// Send error message to client
		http.Error(rw, reqErr.Error(), reqErr.StatusCode)
		return
	}

	http.Error(rw, err.Error(), http.StatusBadRequest)
}

// handleJsonResponse is a helper function for unified JSON response handling.
func handleJsonResponse(rw http.ResponseWriter, status int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
}
