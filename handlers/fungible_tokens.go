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

func (s *FungibleTokens) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Fungible token details")
}

func (s *FungibleTokens) Init(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Init fungible token")
}

func (s *FungibleTokens) ListWithdrawals(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("List withdrawals for fungible token")
}

func (s *FungibleTokens) CreateWithdrawal(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Create withdrawal for fungible token")
}

func (s *FungibleTokens) WithdrawalDetails(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Withdrawal details for fungible token")
}
