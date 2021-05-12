package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
)

type Transactions struct {
	log     *log.Logger
	service *transactions.Service
}

type CreateTransactionBody struct {
	Code      string                        `json:"code"`
	Arguments []transactions.TransactionArg `json:"arguments"`
}

// NewTransactions initiates a new transactions server.
func NewTransactions(l *log.Logger, service *transactions.Service) *Transactions {
	return &Transactions{l, service}
}

func (s *Transactions) List(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := s.service.List(vars["address"])

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Transactions) Create(rw http.ResponseWriter, r *http.Request) {
	var err error

	if r.Body == nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		}
		handleError(rw, s.log, err)
		return
	}

	vars := mux.Vars(r)

	var b CreateTransactionBody

	// Try to decode the request body into the struct.
	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid body"),
		}
		handleError(rw, s.log, err)
		return
	}

	// Decide whether to serve sync or async, default async
	var res interface{}
	if us := r.Header.Get(SYNC_HEADER); us != "" {
		res, err = s.service.CreateSync(r.Context(), b.Code, b.Arguments, vars["address"])
	} else {
		res, err = s.service.CreateAsync(b.Code, b.Arguments, vars["address"])
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *Transactions) Details(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := s.service.Details(vars["address"], vars["transactionId"])

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
