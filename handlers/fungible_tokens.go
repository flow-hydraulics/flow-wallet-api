package handlers

import (
	"log"
	"net/http"
)

type FungibleTokens struct {
	log *log.Logger
}

func NewFungibleTokens(l *log.Logger) *FungibleTokens {
	return &FungibleTokens{l}
}

func (s *FungibleTokens) Details(rw http.ResponseWriter, r *http.Request) {}

func (s *FungibleTokens) Init(rw http.ResponseWriter, r *http.Request) {}

func (s *FungibleTokens) ListWithdrawals(rw http.ResponseWriter, r *http.Request) {}

func (s *FungibleTokens) CreateWithdrawal(rw http.ResponseWriter, r *http.Request) {}

func (s *FungibleTokens) WithdrawalDetails(rw http.ResponseWriter, r *http.Request) {}
