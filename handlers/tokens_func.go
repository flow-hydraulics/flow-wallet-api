package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/gorilla/mux"
)

func (s *Tokens) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

	res, err := s.service.Details(r.Context(), tokenName, address)
	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) CreateWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

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
	job, tx, err := s.service.CreateFtWithdrawal(r.Context(), sync, tokenName, address, b.Recipient, b.Amount)
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

func (s *Tokens) ListWithdrawalsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

	res, err := s.service.ListFtWithdrawals(address, tokenName)
	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) GetWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]
	txId := vars["transactionId"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

	res, err := s.service.GetFtWithdrawal(address, tokenName, txId)
	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) ListDepositsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

	res, err := s.service.ListFtDeposits(address, tokenName)
	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) GetDepositFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]
	transactionId := vars["transactionId"]

	if s.tokenType == templates.NFT {
		handleError(rw, s.log, fmt.Errorf("not yet implemented"))
		return
	}

	res, err := s.service.GetFtDeposit(address, tokenName, transactionId)
	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
