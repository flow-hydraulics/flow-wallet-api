package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/gorilla/mux"
)

func (s *FungibleTokens) CreateWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		handleError(rw, s.log, &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		})
		return
	}

	var b FTWithdrawalRequest

	// Try to decode the request body into the struct.
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		handleError(rw, s.log, &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid body"),
		})
		return
	}

	vars := mux.Vars(r)
	sender := vars["address"]
	token := vars["tokenName"]

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SYNC_HEADER) != ""
	job, t, err := s.service.CreateFtWithdrawal(r.Context(), sync, token, sender, b.Recipient, b.Amount)
	var res interface{}
	if sync {
		res = t
	} else {
		res = job
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *FungibleTokens) SetupFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	account := vars["address"]
	token := vars["tokenName"]

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SYNC_HEADER) != ""
	job, t, err := s.service.SetupFtForAccount(r.Context(), sync, token, account)
	var res interface{}
	if sync {
		res = t
	} else {
		res = job
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}
