package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk/client"
)

type Transactions struct {
	l  *log.Logger
	c  *client.Client
	db store.DataStore
	ks keys.KeyStore
}

func NewTransactions(
	l *log.Logger,
	c *client.Client,
	db store.DataStore,
	ks keys.KeyStore) *Transactions {
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
