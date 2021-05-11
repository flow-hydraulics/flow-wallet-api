package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
)

type Transactions struct {
	log     *log.Logger
	service *transactions.Service
}

// NewTransactions initiates a new transactions server.
func NewTransactions(l *log.Logger, service *transactions.Service) *Transactions {
	return &Transactions{l, service}
}

func (s *Transactions) List(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("List transactions")
	vars := mux.Vars(r)
	result, err := s.service.List(vars["address"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(result)
}

func (s *Transactions) Create(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Create transaction")
	vars := mux.Vars(r)
	var t transactions.Transaction
	// Try to decode the request body into the struct.
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	result, err := s.service.CreateSync(r.Context(), t.Code, t.Arguments, vars["address"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(result)
}

func (s *Transactions) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Transaction details")
	vars := mux.Vars(r)
	result, err := s.service.Details(vars["address"], vars["transactionId"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(result)
}
