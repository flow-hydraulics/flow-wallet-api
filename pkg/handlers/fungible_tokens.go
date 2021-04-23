package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk/client"
)

type FungibleTokens Server

func NewFungibleTokens(
	l *log.Logger,
	c *client.Client,
	db store.DataStore,
	ks store.KeyStore) *FungibleTokens {
	return &FungibleTokens{l, c, db, ks}
}

func (s *FungibleTokens) Details(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Fungible token details")
}

func (s *FungibleTokens) Init(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Init fungible token")
}

func (s *FungibleTokens) ListWithdrawals(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("List withdrawals for fungible token")
}

func (s *FungibleTokens) CreateWithdrawal(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Create withdrawal for fungible token")
}

func (s *FungibleTokens) WithdrawalDetails(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Withdrawal details for fungible token")
}
