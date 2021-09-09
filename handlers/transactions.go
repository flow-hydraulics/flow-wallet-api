package handlers

import (
	"log"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

type Transactions struct {
	log     *log.Logger
	service *transactions.Service
}

// NewTransactions initiates a new transactions server.
func NewTransactions(l *log.Logger, service *transactions.Service) *Transactions {
	return &Transactions{l, service}
}

func (s *Transactions) List() http.Handler {
	return http.HandlerFunc(s.ListFunc)
}

func (s *Transactions) Create() http.Handler {
	h := http.HandlerFunc(s.CreateFunc)
	return UseJson(h)
}

func (s *Transactions) Sign() http.Handler {
	h := http.HandlerFunc(s.SignFunc)
	return UseJson(h)
}

func (s *Transactions) Details() http.Handler {
	return http.HandlerFunc(s.DetailsFunc)
}

func (s *Transactions) ExecuteScript() http.Handler {
	h := http.HandlerFunc(s.ExecuteScriptFunc)
	return UseJson(h)
}
