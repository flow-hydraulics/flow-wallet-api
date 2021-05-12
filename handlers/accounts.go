package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/gorilla/mux"
)

// Accounts is a HTTP server for account management.
// It provides list, create and details APIs.
// It uses an account service to interface with data.
type Accounts struct {
	log     *log.Logger
	service *accounts.Service
}

// NewAccounts initiates a new accounts server.
func NewAccounts(l *log.Logger, service *accounts.Service) *Accounts {
	return &Accounts{l, service}
}

// List returns all accounts.
func (s *Accounts) List(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("List accounts")
	res, err := s.service.List()
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

// Create creates a new account asynchronously.
// It returns a Job JSON representation.
func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Create account")

	var err error

	// Decide whether to serve sync or async, default async
	var res interface{}
	if us := r.Header.Get(SYNC_HEADER); us != "" {
		res, err = s.service.CreateSync(r.Context())
	} else {
		res, err = s.service.CreateAsync()
	}

	if err != nil {
		handleError(err, s.log, rw)
		return
	}

	handleJsonResponse(rw, http.StatusCreated)
	json.NewEncoder(rw).Encode(res)
}

// Details returns details regarding an account.
// It reads the address for the wanted account from URL.
// Account service is responsible for validating the address.
func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Account details")
	vars := mux.Vars(r)
	res, err := s.service.Details(vars["address"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
