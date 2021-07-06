package accounts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/eqlabs/flow-wallet-api/datastore"
	"github.com/eqlabs/flow-wallet-api/errors"
	"github.com/eqlabs/flow-wallet-api/flow_helpers"
	"github.com/eqlabs/flow-wallet-api/jobs"
	"github.com/eqlabs/flow-wallet-api/keys"
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/eqlabs/flow-wallet-api/transactions"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for account management.
type Service struct {
	store        Store
	km           keys.Manager
	fc           *client.Client
	wp           *jobs.WorkerPool
	transactions *transactions.Service
	templates    *templates.Service
	cfg          Config
}

// NewService initiates a new account service.
func NewService(
	store Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
	txs *transactions.Service,
	tes *templates.Service,
) *Service {
	cfg := ParseConfig()
	return &Service{store, km, fc, wp, txs, tes, cfg}
}

func (s *Service) InitAdminAccount() {
	a, err := s.store.Account(s.cfg.AdminAccountAddress)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			panic(err)
		}
		// Admin account not in database
		a = Account{Address: s.cfg.AdminAccountAddress}
		s.store.InsertAccount(&a)
		// TODO: Enable FlowToken
	}
}

// List returns all accounts in the datastore.
func (s *Service) List(limit, offset int) (result []Account, err error) {
	o := datastore.ParseListOptions(limit, offset)
	return s.store.Accounts(o)
}

// Create calls account.New to generate a new account.
// It receives a new account with a corresponding private key or resource ID
// and stores both in datastore.
// It returns a job, the new account and a possible error.
func (s *Service) Create(c context.Context, sync bool) (*jobs.Job, *Account, error) {
	a := Account{}
	k := keys.Private{}

	job, err := s.wp.AddJob(func() (string, error) {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		if err := New(&a, &k, ctx, s.fc, s.km); err != nil {
			return "", err
		}

		// Convert the key to storable form (encrypt it)
		accountKey, err := s.km.Save(k)
		if err != nil {
			return "", err
		}

		// Store account and key
		a.Keys = []keys.Storable{accountKey}
		if err := s.store.InsertAccount(&a); err != nil {
			return "", err
		}

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

	// TODO: Enable FlowToken

	return job, &a, err
}

// Details returns a specific account.
func (s *Service) Details(address string) (Account, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return Account{}, err
	}

	return s.store.Account(address)
}
