package account

import (
	"context"
	"fmt"
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type Service struct {
	l       *log.Logger
	db      data.Store
	ks      keys.Store
	fc      *client.Client
	chainId flow.ChainID // TODO: how do we want to handle different chains?
}

func NewService(
	l *log.Logger,
	db data.Store,
	ks keys.Store,
	fc *client.Client) *Service {
	return &Service{l, db, ks, fc, flow.Emulator}
}

func (s *Service) List(ctx context.Context) (accounts []data.Account, err error) {
	accounts, err = s.db.Accounts()
	return
}

func (s *Service) Create(ctx context.Context) (account data.Account, err error) {
	account, key, err := Create(ctx, s.fc, s.ks)
	if err != nil {
		return
	}

	// Store the generated key
	err = s.ks.Save(key)
	if err != nil {
		return
	}

	// Store the account
	err = s.db.InsertAccount(account)

	return
}

func (s *Service) Details(ctx context.Context, address string) (account data.Account, err error) {
	err = s.ValidateAddress(address)
	if err != nil {
		return
	}
	account, err = s.db.Account(address)
	return
}

func (s *Service) ValidateAddress(address string) (err error) {
	flowAddress := flow.HexToAddress(address)
	if !flowAddress.IsValid(s.chainId) {
		err = fmt.Errorf("'address': %s is not a valid address in '%s' chain", address, s.chainId)
	}
	return
}
