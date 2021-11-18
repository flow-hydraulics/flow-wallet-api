package handlers

import (
	"net/http"
)

func HandleHealthReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Liveness(getLiveness func() (interface{}, error)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		liveness, err := getLiveness()
		if err != nil {
			handleError(rw, nil, err)
		}
		handleJsonResponse(rw, http.StatusOK, liveness)
	})
}
