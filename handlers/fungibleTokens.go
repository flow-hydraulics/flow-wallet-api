package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/tokens"
)

type FungibleTokens struct {
	log     *log.Logger
	service *tokens.Service
}

type FTWithdrawalRequest struct {
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

func NewFungibleTokens(l *log.Logger, service *tokens.Service) *FungibleTokens {
	return &FungibleTokens{l, service}
}

func (s *FungibleTokens) CreateWithdrawal() http.Handler {
	h := http.HandlerFunc(s.CreateWithdrawalFunc)
	return UseJson(h)
}
