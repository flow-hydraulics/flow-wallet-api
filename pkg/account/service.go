package account

import (
	"context"
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type Service struct {
	l  *log.Logger
	db store.DataStore
	ks store.KeyStore
	fc *client.Client
}

func NewService(
	l *log.Logger,
	db store.DataStore,
	ks store.KeyStore,
	fc *client.Client) *Service {
	return &Service{l, db, ks, fc}
}

func (s *Service) List(context.Context) ([]store.Account, error) {
	accounts, err := s.db.Accounts()
	if err != nil {
		return []store.Account{}, err
	}
	return accounts, nil
}

func (s *Service) Create(ctx context.Context) (store.Account, error) {
	account, key, err := New(ctx, s.fc, s.ks)
	if err != nil {
		return store.Account{}, err
	}

	// Store the generated key
	s.ks.Save(key)

	// Store the account
	s.db.InsertAccount(account)

	return account, nil
}

func (s *Service) Details(ctx context.Context, addr flow.Address) (store.Account, error) {
	account, err := s.db.Account(addr)
	if err != nil {
		return store.Account{}, err
	}
	return account, nil
}
