package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/gorilla/mux"
)

func (s *FungibleTokens) ListFunc(rw http.ResponseWriter, r *http.Request) {
	handleJsonResponse(rw, http.StatusOK, s.service.List())
}

func (s *FungibleTokens) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	a := vars["address"]
	t := vars["tokenName"]

	res, err := s.service.Details(r.Context(), t, a)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *FungibleTokens) CreateFtWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a := vars["address"]
	t := vars["tokenName"]

	var b FTWithdrawalRequest

	if r.Body == nil || r.Body == http.NoBody {
		err := &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("empty body")}
		handleError(rw, s.log, err)
		return
	}

	// Try to decode the request body.
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		err = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}
		handleError(rw, s.log, err)
		return
	}

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SyncHeader) != ""
	job, tx, err := s.service.CreateFtWithdrawal(r.Context(), sync, t, a, b.Recipient, b.Amount)
	var res interface{}
	if sync {
		res = tx
	} else {
		res = job
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *FungibleTokens) ListFtWithdrawalsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a := vars["address"]
	t := vars["tokenName"]

	res, err := s.service.ListFtWithdrawals(a, t)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *FungibleTokens) GetFtWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a := vars["address"]
	t := vars["tokenName"]
	txId := vars["transactionId"]

	res, err := s.service.GetFtWithdrawal(a, t, txId)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *FungibleTokens) ListFtDepositsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a := vars["address"]
	t := vars["tokenName"]

	res, err := s.service.ListFtDeposits(a, t)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *FungibleTokens) GetFtDepositFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a := vars["address"]
	t := vars["tokenName"]
	txId := vars["transactionId"]

	res, err := s.service.GetFtDeposit(a, t, txId)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
