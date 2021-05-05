package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
)

func handleError(err error, logger *log.Logger, rw http.ResponseWriter, r *http.Request) {
	logger.Printf("Error: %v\n", err)
	reqErr, isReqErr := err.(*errors.RequestError)
	if isReqErr {
		http.Error(rw, reqErr.Error(), reqErr.StatusCode)
		return
	}
	http.Error(rw, "Error", http.StatusInternalServerError)
}

func handleJsonOk(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

func handleJsonCreated(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
}
