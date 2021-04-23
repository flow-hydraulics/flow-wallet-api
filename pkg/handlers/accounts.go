package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk/client"
)

type Accounts Server

func NewAccounts(
	l *log.Logger,
	c *client.Client,
	db *store.DataStore,
	ks *store.KeyStore) *Accounts {
	return &Accounts{l, c, db, ks}
}

func (s *Accounts) List(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("List accounts")
}

func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Create account")
}

func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Account details")
}

func (s *Accounts) Update(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Update account")
}

func (s *Accounts) Delete(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Delete account")
}
