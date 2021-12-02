package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/gorilla/mux"
)

func (s *Tokens) SetupFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	// Decide whether to serve sync or async, default async
	sync := r.FormValue(SyncQueryParameter) != ""
	job, transaction, err := s.service.Setup(r.Context(), sync, tokenName, address)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	var res interface{}
	if sync {
		res = transaction.ToJSONResponse()
	} else {
		res = job.ToJSONResponse()
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *Tokens) MakeAccountTokensFunc(tType templates.TokenType) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		a := vars["address"]

		res, err := s.service.AccountTokens(a, tType)

		if err != nil {
			handleError(rw, r, err)
			return
		}

		handleJsonResponse(rw, http.StatusOK, res)
	}
}

func (s *Tokens) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	res, err := s.service.Details(r.Context(), tokenName, address)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) CreateWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	var withdrawal tokens.WithdrawalRequest

	if r.Body == nil || r.Body == http.NoBody {
		err := &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("empty body")}
		handleError(rw, r, err)
		return
	}

	// Try to decode the request body.
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		err = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}
		handleError(rw, r, err)
		return
	}

	withdrawal.TokenName = tokenName

	// Decide whether to serve sync or async, default async
	sync := r.FormValue(SyncQueryParameter) != ""
	job, transaction, err := s.service.CreateWithdrawal(r.Context(), sync, address, withdrawal)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	var res interface{}
	if sync {
		res = transaction.ToJSONResponse()
	} else {
		res = job.ToJSONResponse()
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *Tokens) ListWithdrawalsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	res, err := s.service.ListWithdrawals(address, tokenName)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) GetWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]
	txId := vars["transactionId"]

	res, err := s.service.GetWithdrawal(address, tokenName, txId)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) ListDepositsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]

	res, err := s.service.ListDeposits(address, tokenName)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Tokens) GetDepositFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	tokenName := vars["tokenName"]
	transactionId := vars["transactionId"]

	res, err := s.service.GetDeposit(address, tokenName, transactionId)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
