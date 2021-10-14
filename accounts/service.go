package accounts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
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
	cfg          *configs.Config
}

// NewService initiates a new account service.
func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
	txs *transactions.Service,
	tes *templates.Service,
) *Service {
	// TODO(latenssi): safeguard against nil config?
	return &Service{store, km, fc, wp, txs, tes, cfg}
}

func (s *Service) InitAdminAccount(ctx context.Context) error {
	a, err := s.store.Account(s.cfg.AdminAddress)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		// Admin account not in database
		a = Account{Address: s.cfg.AdminAddress}
		err := s.store.InsertAccount(&a)
		if err != nil {
			return err
		}
		AccountAdded.Trigger(AccountAddedPayload{
			Address: flow.HexToAddress(s.cfg.AdminAddress),
		})
	}

	keyCount, err := s.km.InitAdminProposalKeys(ctx)
	if err != nil {
		return err
	}

	if keyCount < s.cfg.AdminProposalKeyCount {
		err = s.addAdminProposalKeys(ctx, s.cfg.AdminProposalKeyCount-keyCount)
		if err != nil {
			return err
		}

		_, err = s.km.InitAdminProposalKeys(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) addAdminProposalKeys(ctx context.Context, count uint16) error {
	_, _, err := s.transactions.Create(ctx, true, s.cfg.AdminAddress, templates.Raw{
		Code: template_strings.AddProposalKeyTransaction,
		Arguments: []templates.Argument{
			cadence.NewInt(s.cfg.AdminKeyIndex),
			cadence.NewUInt16(count),
		},
	}, transactions.General)

	return err
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
	a := &Account{Type: AccountTypeCustodial}
	k := &keys.Private{}
	transaction := &transactions.Transaction{}

	process := func(jobResult *jobs.Result) error {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		err := New(ctx, transaction, a, k, s.fc, s.km, s.cfg.TransactionTimeout)
		jobResult.TransactionID = transaction.TransactionId // Update job result
		if err != nil {
			return err
		}

		// Convert the key to storable form (encrypt it)
		accountKey, err := s.km.Save(*k)
		if err != nil {
			return err
		}

		// Store account and key
		a.Keys = []keys.Storable{accountKey}
		if err := s.store.InsertAccount(a); err != nil {
			return err
		}

		AccountAdded.Trigger(AccountAddedPayload{
			Address: flow.HexToAddress(a.Address),
		})

		// Update job result
		jobResult.Result = a.Address

		return nil
	}

	job, err := s.wp.AddJob(process)

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

	return job, a, err
}

func (s *Service) AddNonCustodialAccount(_ context.Context, address string) (*Account, error) {
	a := &Account{
		Address: flow_helpers.HexString(address),
		Type:    AccountTypeNonCustodial,
	}

	err := s.store.InsertAccount(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Service) DeleteNonCustodialAccount(_ context.Context, address string) error {
	a, err := s.store.Account(flow_helpers.HexString(address))
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Account already gone. All good.
			return nil
		}

		return err
	}

	if a.Type != AccountTypeNonCustodial {
		return fmt.Errorf("only non-custodial accounts supported")
	}

	return s.store.HardDeleteAccount(&a)
}

// Details returns a specific account.
func (s *Service) Details(address string) (Account, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return Account{}, err
	}

	return s.store.Account(address)
}
