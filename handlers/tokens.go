package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/tokens"
)

type Tokens struct {
	log       *log.Logger
	service   *tokens.Service
	tokenType templates.TokenType
}

type FTWithdrawalRequest struct {
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

func NewTokens(l *log.Logger, service *tokens.Service, tType templates.TokenType) *Tokens {
	return &Tokens{l, service, tType}
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
