package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/account"
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
	accounts, err := s.as.List(context.Background())
	if err != nil {
		s.l.Printf("Error: %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error"))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(accounts)
}

func (s *Accounts) Create(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Create account")
	account, err := s.as.Create(context.Background())
	if err != nil {
		s.l.Printf("Error: %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error"))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(account)
}

func (s *Accounts) Details(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Account details")
	vars := mux.Vars(r)
	account, err := s.as.Details(context.Background(), vars["address"])
	if err != nil {
		s.l.Printf("Error: %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error"))
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(account)

}
