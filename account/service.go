package account

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// Datastore is the interface required by account service for data storage.
type Datastore interface {
	Accounts() ([]data.Account, error)
	InsertAccount(a data.Account) error
	Account(address string) (data.Account, error)
}

// Service defines the API for account management.
type Service struct {
	l   *log.Logger
	db  Datastore
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
	cfg Config
}

// NewService initiates a new account service.
func NewService(
	l *log.Logger,
	db Datastore,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	cfg := ParseConfig()
	return &Service{l, db, km, fc, wp, cfg}
}

// List returns all accounts in the datastore.
func (s *Service) List() (result []data.Account, err error) {
	return s.db.Accounts()
}

// Create calls account.Create to generate a new account.
// It receives a new account with a corresponding private key or resource ID
// and stores both in datastore.
// It fetches and returns the new account from datastore.
func (s *Service) Create(ctx context.Context) (result data.Account, err error) {
	account, key, err := Create(ctx, s.fc, s.km)
	if err != nil {
		return
	}

	// Convert the key to storable form (encrypt it)
	accountKey, err := s.km.Save(key)
	if err != nil {
		return
	}

	// Store account and key
	account.Keys = []data.Key{accountKey}
	err = s.db.InsertAccount(account)
	if err != nil {
		return
	}

	// Need to get from database to populate `CreatedAt` and `UpdatedAt` fields
	return s.db.Account(account.Address)
}

// CreateAsync calls service.Create asynchronously.
// It creates a job and returns it. This allows us to respond with a job id
// which the client can use to poll for the results later.
func (s *Service) CreateAsync() (*jobs.Job, error) {
	job, err := s.wp.AddJob(func() (string, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		account, err := s.Create(ctx)
		if err != nil {
			s.l.Println(err)
			return "", err
		}
		return account.Address, nil
	})

	if err != nil {
		_, isJErr := err.(*errors.JobQueueFull)
		if isJErr {
			err = &errors.RequestError{
				StatusCode: http.StatusServiceUnavailable,
				Err:        fmt.Errorf("max capacity reached, try again later"),
			}
		}
		return nil, err
	}

	return job, nil
}

// Details returns a specific account.
func (s *Service) Details(address string) (result data.Account, err error) {
	// Check if the input is a valid address
	err = s.ValidateAddress(address)
	if err != nil {
		return
	}

	// Get from datastore
	result, err = s.db.Account(address)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("account not found"),
		}
	}

	return
}

// ValidateAddress checks if the given address is valid in the current Flow network.
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
