package accounts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

func (s *Service) InitAdminAccount() {
	a, err := s.db.Account(s.cfg.AdminAccountAddress)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			panic(err)
		}
		// Admin account not in database
		a = Account{Address: s.cfg.AdminAccountAddress}
		s.db.InsertAccount(&a)
	}

	for _, t := range templates.EnabledTokens() {
		at := AccountToken{
			AccountAddress: a.Address,
			TokenAddress:   t.Address,
			TokenName:      t.CanonName(),
		}
		s.db.InsertAccountToken(&at) // Ignore errors
	}
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
		if err := s.db.InsertAccount(&a); err != nil {
			return "", err
		}

		// Store an AccountToken named FlowToken for the account as it is automatically
		// enabled on all accounts
		t, err := templates.NewToken("FlowToken")
		if err == nil { // If err != nil, FlowToken is not enabled for some reason
			s.db.InsertAccountToken(&AccountToken{ // Ignore errors
				AccountAddress: a.Address,
				TokenAddress:   t.Address,
				TokenName:      t.CanonName(),
			})
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

	return job, &a, err
}

// Details returns a specific account.
func (s *Service) Details(address string) (Account, error) {
	// Check if the input is a valid address
	if err := flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
		return Account{}, err
	}
	address = flow_helpers.HexString(address)

	return s.db.Account(address)
}

func (s *Service) SetupFungibleToken(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	if err := flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
		return nil, nil, err
	}

	address = flow_helpers.HexString(address)

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return nil, nil, err
	}

	raw := templates.Raw{
		Code: templates.FungibleSetupCode(token),
	}

	job, tx, err := s.ts.Create(ctx, sync, address, raw, transactions.FtSetup)

	// Handle adding token to account in database
	go func() {
		if err := job.Wait(true); err != nil && !strings.Contains(err.Error(), "vault exists") {
			return
		}

		err = s.db.InsertAccountToken(&AccountToken{
			AccountAddress: address,
			TokenAddress:   token.Address,
			TokenName:      token.CanonName(),
		})

		if err != nil && !strings.Contains(err.Error(), "duplicate key value") {
			fmt.Printf("error while adding account token: %s\n", err)
		}
	}()

	return job, tx, err
}

func (s *Service) AccountFungibleTokens(address string) ([]AccountToken, error) {
	// Check if the input is a valid address
	if err := flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
		return nil, err
	}

	address = flow_helpers.HexString(address)

	tt, err := s.db.AccountTokens(address)
	if err != nil {
		return []AccountToken{}, err
	}

	return tt, nil
}
