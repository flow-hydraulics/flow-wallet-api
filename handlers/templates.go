package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/templates"
)

// Accounts is a HTTP server for account management.
// It provides list, create and details APIs.
// It uses an account service to interface with data.
type Templates struct {
	log     *log.Logger
	service *templates.Service
}

// NewAccounts initiates a new accounts server.
func NewTemplates(l *log.Logger, service *templates.Service) *Templates {
	return &Templates{l, service}
}

func (s *Templates) AddToken() http.Handler {
	h := http.HandlerFunc(s.AddTokenFunc)
	return UseJson(h)
}

func (s *Templates) ListTokens() http.Handler {
	return http.HandlerFunc(s.ListTokensFunc)
}

func (s *Templates) GetToken() http.Handler {
	return http.HandlerFunc(s.GetTokenFunc)
}

func (s *Templates) RemoveToken() http.Handler {
	return http.HandlerFunc(s.RemoveTokenFunc)
}
