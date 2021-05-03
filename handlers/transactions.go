package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk/client"
)

type Transactions struct {
	l  *log.Logger
	c  *client.Client
	db data.Store
	km keys.Manager
}

func NewTransactions(
	l *log.Logger,
	c *client.Client,
	db data.Store,
	km keys.Manager) *Transactions {
	return &Transactions{l, c, db, km}
}

func (s *Transactions) List(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("List transactions")
}

func (s *Transactions) Create(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Create transaction")
}

func (s *Transactions) Details(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Transaction details")
}
