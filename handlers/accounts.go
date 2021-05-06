package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/account"
	"github.com/gorilla/mux"
)

// Accounts is a HTTP server for account management.
// It provides list, create and details APIs.
// It uses an account service to interface with data.
type Accounts struct {
	log     *log.Logger
	service *account.Service
}

// NewAccounts initiates a new accounts server.
func NewAccounts(log *log.Logger, service *account.Service) *Accounts {
	return &Accounts{log, service}
}

// List returns all accounts.
func (s *Accounts) List(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("List accounts")
	result, err := s.service.List(r.Context())
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(result)
}

// Create creates a new account.
func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Create account")
	result, err := s.service.Create(r.Context())
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusCreated)
	json.NewEncoder(rw).Encode(result)
}

// Details returns details regarding an account.
// It reads the address for the wanted account from URL.
// Account service is responsible for validating the address.
func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Account details")
	vars := mux.Vars(r)
	result, err := s.service.Details(r.Context(), vars["address"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(result)
}
