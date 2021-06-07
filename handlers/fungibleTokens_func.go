package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/gorilla/mux"
)

func (s *FungibleTokens) CreateWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sender := vars["address"]
	token := vars["tokenName"]

	var b FTWithdrawalRequest
	var res interface{}

	if r.Body == nil {
		err := &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("empty body")}
		handleError(rw, s.log, err)
		return
	}

	// Try to decode the request body into the struct.
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		err = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}
		handleError(rw, s.log, err)
		return
	}

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SYNC_HEADER) != ""
	job, t, err := s.service.CreateFtWithdrawal(r.Context(), sync, token, sender, b.Recipient, b.Amount)

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

	var b FTSetupRequest
	var res interface{}

	if r.Body != nil {
		// Try to decode the request body into the struct.
		err := json.NewDecoder(r.Body).Decode(&b)
		if err != nil {
			err = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}
			handleError(rw, s.log, err)
			return
		}
	}

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SYNC_HEADER) != ""
	job, t, err := s.service.SetupFtForAccount(r.Context(), sync, token, account, b.TokenAddress)
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
