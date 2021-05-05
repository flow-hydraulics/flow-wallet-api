package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/account"
	"github.com/gorilla/mux"
)

type Accounts struct {
	l  *log.Logger
	as *account.Service
}

func NewAccounts(l *log.Logger, as *account.Service) *Accounts {
	return &Accounts{l, as}
}

func (s *Accounts) List(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("List accounts")
	result, err := s.as.List(context.Background())
	if err != nil {
		handleError(err, s.l, rw, r)
		return
	}
	handleJsonOk(rw)
	json.NewEncoder(rw).Encode(result)
}

func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Create account")
	result, err := s.as.Create(context.Background())
	if err != nil {
		handleError(err, s.l, rw, r)
		return
	}
	handleJsonCreated(rw)
	json.NewEncoder(rw).Encode(result)
}

func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Account details")
	vars := mux.Vars(r)
	result, err := s.as.Details(context.Background(), vars["address"])
	if err != nil {
		handleError(err, s.l, rw, r)
		return
	}
	handleJsonOk(rw)
	json.NewEncoder(rw).Encode(result)
}
