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
	s.log.Println("List transactions")

	vars := mux.Vars(r)

	res, err := s.service.List(vars["address"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}

	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (s *Transactions) Create(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Create transaction")

	var err error

	if r.Body == nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		}
		handleError(err, s.log, rw)
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
		handleError(err, s.log, rw)
		return
	}

	// Decide whether to serve sync or async, default async
	var res interface{}
	if us := r.Header.Get("Use-Sync"); us != "" {
		res, err = s.service.CreateSync(r.Context(), b.Code, b.Arguments, vars["address"])
	} else {
		res, err = s.service.CreateAsync(b.Code, b.Arguments, vars["address"])
	}

	if err != nil {
		handleError(err, s.log, rw)
		return
	}

	handleJsonResponse(rw, http.StatusCreated)
	json.NewEncoder(rw).Encode(res)
}

func (s *Transactions) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Transaction details")

	vars := mux.Vars(r)

	res, err := s.service.Details(vars["address"], vars["transactionId"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}

	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
