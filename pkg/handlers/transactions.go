package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk/client"
)

type Transactions struct {
	l  *log.Logger
	c  *client.Client
	db data.Store
	ks keys.Store
}

func NewTransactions(
	l *log.Logger,
	c *client.Client,
	db data.Store,
	ks keys.Store) *Transactions {
	return &Transactions{l, c, db, ks}
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
