package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-api/tokens"
)

type Tokens struct {
	log     *log.Logger
	service *tokens.Service
}

func NewTokens(l *log.Logger, service *tokens.Service) *Tokens {
	return &Tokens{l, service}
}

func (s *Tokens) Details() http.Handler {
	h := http.HandlerFunc(s.DetailsFunc)
	return h
}

func (s *Tokens) CreateWithdrawal() http.Handler {
	h := http.HandlerFunc(s.CreateWithdrawalFunc)
	return UseJson(h)
}

func (s *Tokens) ListWithdrawals() http.Handler {
	h := http.HandlerFunc(s.ListWithdrawalsFunc)
	return h
}

func (s *Tokens) GetWithdrawal() http.Handler {
	h := http.HandlerFunc(s.GetWithdrawalFunc)
	return h
}

func (s *Tokens) ListDeposits() http.Handler {
	h := http.HandlerFunc(s.ListDepositsFunc)
	return h
}

func (s *Tokens) GetDeposit() http.Handler {
	h := http.HandlerFunc(s.GetDepositFunc)
	return h
}
