// Package handlers provides HTTP handlers for different services across the application.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gorilla "github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/handlers/middleware"
)

const SyncQueryParameter = "sync"

var EmptyBodyError = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("empty body")}
var InvalidBodyError = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}

func UseCors(h http.Handler) http.Handler {
	return gorilla.CORS(gorilla.AllowedOrigins([]string{"*"}))(h)
}

func UseLogging(h http.Handler) http.Handler {
	return middleware.LoggingHandler(h)
}

func UseCompress(h http.Handler) http.Handler {
	return gorilla.CompressHandler(h)
}

func UseJson(h http.Handler) http.Handler {
	// Only PUT, POST, and PATCH requests are considered.
	return gorilla.ContentTypeHandler(h, "application/json")
}

func UseIdempotency(h http.Handler, opts IdempotencyHandlerOptions, store IdempotencyStore) http.Handler {
	return IdempotencyHandler(h, opts, store)
}

// handleError is a helper function for unified HTTP error handling.
func handleError(rw http.ResponseWriter, r *http.Request, err error) {
	log.
		WithFields(log.Fields{"error": err}).
		Warn("Error while handling request")

		// Check if the error was an errors.RequestError
	reqErr, isReqErr := err.(*errors.RequestError)
	if isReqErr {
		http.Error(rw, reqErr.Error(), reqErr.StatusCode)
		return
	}

	// Check for "record not found" database error
	if strings.Contains(err.Error(), "record not found") {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(rw, err.Error(), http.StatusBadRequest)
}

// handleJsonResponse is a helper function for unified JSON response handling.
func handleJsonResponse(rw http.ResponseWriter, status int, res interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(res) // nolint
}

func checkNonEmptyBody(r *http.Request) error {
	if r.Body == nil || r.Body == http.NoBody {
		return EmptyBodyError
	}
	return nil
}

func servePlainText(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s)) // nolint
}
