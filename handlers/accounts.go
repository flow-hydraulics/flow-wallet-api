package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/templates"
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

func (s *Accounts) Details() http.Handler {
	return http.HandlerFunc(s.DetailsFunc)
}

func (s *Accounts) SetupToken() http.Handler {
	h := http.HandlerFunc(s.SetupTokenFunc)
	return UseJson(h)
}

func (s *Accounts) AccountTokens(tType templates.TokenType) http.Handler {
	return s.MakeAccountTokensFunc(tType)
}
