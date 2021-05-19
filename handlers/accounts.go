package handlers

import (
	"log"
	"net/http"
	"strconv"

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
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 0
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		offset = 0
	}

	res, err := s.service.List(limit, offset)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

// Create creates a new account asynchronously.
// It returns a Job JSON representation.
func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	var err error

	// Decide whether to serve sync or async, default async
	var res interface{}
	if us := r.Header.Get(SYNC_HEADER); us != "" {
		res, err = s.service.CreateSync(r.Context())
	} else {
		res, err = s.service.CreateAsync()
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

// Details returns details regarding an account.
// It reads the address for the wanted account from URL.
// Account service is responsible for validating the address.
func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := s.service.Details(vars["address"])

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
