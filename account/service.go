package account

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type Datastore interface {
	Accounts() ([]data.Account, error)
	InsertAccount(a data.Account) error
	Account(address string) (data.Account, error)
}

type Service struct {
	l   *log.Logger
	db  Datastore
	km  keys.Manager
	fc  *client.Client
	cfg Config
}

func NewService(
	l *log.Logger,
	db Datastore,
	km keys.Manager,
	fc *client.Client,
) *Service {
	cfg := ParseConfig()
	return &Service{l, db, km, fc, cfg}
}

func (s *Service) List(ctx context.Context) (accounts []data.Account, err error) {
	accounts, err = s.db.Accounts()
	return
}

func (s *Service) Create(ctx context.Context) (account data.Account, err error) {
	account, key, err := Create(ctx, s.fc, s.km)
	if err != nil {
		return
	}

	accountKey, err := s.km.Save(key)
	if err != nil {
		return
	}
	account.Keys = []data.Key{accountKey}

	// Store
	err = s.db.InsertAccount(account)
	if err != nil {
		return
	}

	// Need to get from database to populate `CreatedAt` and `UpdatedAt` fields
	account, err = s.db.Account(account.Address)

	return
}

func (s *Service) Details(ctx context.Context, address string) (account data.Account, err error) {
	err = s.ValidateAddress(address)
	if err != nil {
		return
	}

	account, err = s.db.Account(address)
	if err != nil {
		if err.Error() == "record not found" {
			err = &errors.RequestError{
				StatusCode: http.StatusNotFound,
				Err:        fmt.Errorf("account not found"),
			}
		}
	}

	return
}

func (s *Service) ValidateAddress(address string) (err error) {
	flowAddress := flow.HexToAddress(address)
	if !flowAddress.IsValid(s.cfg.ChainId) {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("not a valid address"),
		}
	}
	return
}
