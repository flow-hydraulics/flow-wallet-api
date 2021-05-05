package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
)

func handleError(err error, logger *log.Logger, rw http.ResponseWriter, r *http.Request) {
	if err != nil {
		logger.Printf("Error: %v\n", err)
		e, isReqErr := err.(*errors.RequestError)
		if isReqErr {
			rw.WriteHeader(e.StatusCode)
			rw.Write([]byte(e.Error()))
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error"))
		return
	}
}

func handleJsonOk(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
}

func handleJsonCreated(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
}
