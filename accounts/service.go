package accounts

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eqlabs/flow-wallet-service/datastore"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for account management.
type Service struct {
	db  Store
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
	cfg Config
}

// NewService initiates a new account service.
func NewService(
	db Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	cfg := ParseConfig()
	return &Service{db, km, fc, wp, cfg}
}

// List returns all accounts in the datastore.
func (s *Service) List(limit, offset int) (result []Account, err error) {
	o := datastore.ParseListOptions(limit, offset)
	return s.db.Accounts(o)
}

// Create calls account.New to generate a new account.
// It receives a new account with a corresponding private key or resource ID
// and stores both in datastore.
func (s *Service) create(ctx context.Context) (*Account, error) {
	account, key, err := New(ctx, s.fc, s.km)
	if err != nil {
		return nil, err
	}

	// Convert the key to storable form (encrypt it)
	accountKey, err := s.km.Save(key)
	if err != nil {
		return nil, err
	}

	// Store account and key
	account.Keys = []keys.Storable{accountKey}
	err = s.db.InsertAccount(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// CreateSync calls service.create asynchronously.
// It creates a job but waits for its completion. This allows synchronous
// request/response while still using job queueing to allow only one admin key.
// It returns the new account.
func (s *Service) CreateSync(ctx context.Context) (*Account, error) {
	var result *Account
	var jobErr error
	var createErr error
	var done bool = false

	_, jobErr = s.wp.AddJob(func() (string, error) {
		result, createErr = s.create(context.Background())
		done = true
		if createErr != nil {
			return "", createErr
		}
		return result.Address, nil
	})

	if jobErr != nil {
		_, isJErr := jobErr.(*errors.JobQueueFull)
		if isJErr {
			jobErr = &errors.RequestError{
				StatusCode: http.StatusServiceUnavailable,
				Err:        fmt.Errorf("max capacity reached, try again later"),
			}
		}
		return nil, jobErr
	}

	// Wait for the job to have finished
	for !done {
		time.Sleep(10 * time.Millisecond)
	}

	return result, createErr
}

// CreateAsync calls service.create asynchronously.
// It creates a job and returns it. This allows us to respond with a job id
// which the client can use to poll for the results later.
func (s *Service) CreateAsync() (*jobs.Job, error) {
	job, err := s.wp.AddJob(func() (string, error) {
		account, err := s.create(context.Background())
		if err != nil {
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
func (s *Service) Details(address string) (result Account, err error) {
	// Check if the input is a valid address
	err = ValidateAddress(address, s.cfg.ChainId)
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
func ValidateAddress(address string, chainId flow.ChainID) (err error) {
	flowAddress := flow.HexToAddress(address)
	if !flowAddress.IsValid(chainId) {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("not a valid address"),
		}
	}

	return
}
