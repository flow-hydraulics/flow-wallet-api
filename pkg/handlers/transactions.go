package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk/client"
)

type Transactions Server

func NewTransactions(
	l *log.Logger,
	c *client.Client,
	db store.DataStore,
	ks store.KeyStore) *Transactions {
	return &Transactions{l, c, db, ks}
}

func (s *Transactions) SendTransaction(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Send transaction")
}
