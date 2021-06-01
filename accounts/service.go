package accounts

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eqlabs/flow-wallet-service/datastore"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
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
// It returns a job, the new account and a possible error.
func (s *Service) Create(ctx context.Context, sync bool) (*jobs.Job, *Account, error) {
	var account *Account

	job, err := s.wp.AddJob(func() (string, error) {
		c := ctx
		if !sync {
			c = context.Background()
		}

		a, key, err := New(c, s.fc, s.km)
		if err != nil {
			return "", err
		}

		// Convert the key to storable form (encrypt it)
		accountKey, err := s.km.Save(key)
		if err != nil {
			return "", err
		}

		// Store account and key
		a.Keys = []keys.Storable{accountKey}
		err = s.db.InsertAccount(&a)
		if err != nil {
			return "", err
		}

		account = &a

		return a.Address, nil
	})

	if err != nil {
		_, isJErr := err.(*errors.JobQueueFull)
		if isJErr {
			err = &errors.RequestError{
				StatusCode: http.StatusServiceUnavailable,
				Err:        fmt.Errorf("max capacity reached, try again later"),
			}
		}
		return nil, nil, err
	}

	if sync {
		// Wait for the job to have finished
		for job.Status == jobs.Accepted {
			time.Sleep(10 * time.Millisecond)
		}
		if job.Status == jobs.Error {
			return nil, nil, fmt.Errorf(job.Result)
		}
	}

	return job, account, nil
}

// Details returns a specific account.
func (s *Service) Details(address string) (result Account, err error) {
	// Check if the input is a valid address
	err = flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return
	}

	// Get from datastore
	result, err = s.db.Account(flow_helpers.HexString(address))
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("account not found"),
		}
	}

	return
}
