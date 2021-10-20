package handlers

import (
	"log"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
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

func (s *Accounts) List() http.Handler {
	return http.HandlerFunc(s.ListFunc)
}

func (s *Accounts) Create() http.Handler {
	return http.HandlerFunc(s.CreateFunc)
}

func (s *Accounts) AddNonCustodialAccount() http.Handler {
	return http.HandlerFunc(s.AddNonCustodialAccountFunc)
}

func (s *Accounts) DeleteNonCustodialAccount() http.Handler {
	return http.HandlerFunc(s.DeleteNonCustodialAccountFunc)
}

func (s *Accounts) Details() http.Handler {
	return http.HandlerFunc(s.DetailsFunc)
}
