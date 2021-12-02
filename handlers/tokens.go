package handlers

import (
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	log "github.com/sirupsen/logrus"
)

type Tokens struct {
	logger  *log.Entry
	service *tokens.Service
}

func NewTokens(l *log.Entry, service *tokens.Service) *Tokens {
	return &Tokens{l, service}
}

func (s *Tokens) Setup() http.Handler {
	h := http.HandlerFunc(s.SetupFunc)
	return h
}

func (s *Tokens) AccountTokens(tType templates.TokenType) http.Handler {
	return s.MakeAccountTokensFunc(tType)
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
