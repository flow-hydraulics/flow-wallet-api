package account

import (
	"context"
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk/client"
)

type Service struct {
	l  *log.Logger
	db data.Store
	ks keys.Store
	fc *client.Client
}

func NewService(
	l *log.Logger,
	db data.Store,
	ks keys.Store,
	fc *client.Client) *Service {
	return &Service{l, db, ks, fc}
}

func (s *Service) List(context.Context) ([]data.Account, error) {
	accounts, err := s.db.Accounts()
	if err != nil {
		return []data.Account{}, err
	}
	return accounts, nil
}

func (s *Service) Create(ctx context.Context) (data.Account, error) {
	account, key, err := Create(ctx, s.fc, s.ks)
	if err != nil {
		return data.Account{}, err
	}

	// Store the generated key
	s.ks.Save(key)

	// Store the account
	s.db.InsertAccount(account)

	return account, nil
}

func (s *Service) Details(ctx context.Context, address string) (data.Account, error) {
	account, err := s.db.Account(address)
	if err != nil {
		return data.Account{}, err
	}
	return account, nil
}
