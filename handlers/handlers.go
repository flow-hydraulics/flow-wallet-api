// Package handlers provides HTTP handlers for different services across the application.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/eqlabs/flow-wallet-service/errors"
)

const SyncHeader = "Use-Sync"

// handleError is a helper function for unified HTTP error handling.
func handleError(rw http.ResponseWriter, logger *log.Logger, err error) {
	if logger != nil {
		logger.Printf("Error: %v\n", err)
	}

	// Check if the error was an errors.RequestError
	reqErr, isReqErr := err.(*errors.RequestError)
	if isReqErr {
		http.Error(rw, reqErr.Error(), reqErr.StatusCode)
		return
	}

	// Check for "record not found" database error
	if strings.Contains(err.Error(), "record not found") {
		http.Error(rw, "record not found", http.StatusNotFound)
		return
	}

	http.Error(rw, err.Error(), http.StatusBadRequest)
}

// handleJsonResponse is a helper function for unified JSON response handling.
func handleJsonResponse(rw http.ResponseWriter, status int, res interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(res)
}
