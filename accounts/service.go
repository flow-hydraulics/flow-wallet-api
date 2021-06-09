package accounts

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/datastore"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for account management.
type Service struct {
	db  Store
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
	ts  *transactions.Service
	cfg Config
}

// NewService initiates a new account service.
func NewService(
	db Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
	ts *transactions.Service,
) *Service {
	cfg := ParseConfig()
	return &Service{db, km, fc, wp, ts, cfg}
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
func (s *Service) Create(c context.Context, sync bool) (*jobs.Job, *Account, error) {
	var account *Account

	job, err := s.wp.AddJob(func() (string, error) {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		a, key, err := New(ctx, s.fc, s.km)
		if err != nil {
			return "", err
		}

		account = &a

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

		// Store an AccountToken named FlowToken for the account as it is automatically
		// enabled on all accounts
		// Intentionally ignore error
		s.db.InsertAccountToken(&AccountToken{
			AccountAddress: a.Address,
			Name:           "FlowToken",
		})

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

	err = job.Wait(sync)

	return job, account, err
}

// Details returns a specific account.
func (s *Service) Details(address string) (result Account, err error) {
	// Check if the input is a valid address
	err = flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return
	}
	address = flow_helpers.HexString(address)

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

func (s *Service) SetupFungibleToken(ctx context.Context, sync bool, token templates.Token, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	address = flow_helpers.HexString(address)

	raw := templates.Raw{
		Code: templates.FungibleSetupCode(token, s.cfg.ChainId),
	}

	job, tx, err := s.ts.Create(ctx, sync, address, raw, transactions.FtSetup)

	// Handle adding token to account in database
	go func() {
		err := job.Wait(true)
		if err != nil {
			return
		}
		// Intentionally ignore error
		s.db.InsertAccountToken(&AccountToken{
			AccountAddress: address,
			Name:           token.CanonName(),
		})
	}()

	return job, tx, err
}

func (s *Service) AccountFungibleTokens(address string) ([]AccountToken, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	address = flow_helpers.HexString(address)

	return s.db.AccountTokens(address)
}
