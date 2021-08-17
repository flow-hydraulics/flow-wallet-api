package transactions

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for transaction HTTP handlers.
type Service struct {
	store Store
	km    keys.Manager
	fc    *client.Client
	wp    *jobs.WorkerPool
	cfg   Config
}

// NewService initiates a new transaction service.
func NewService(
	store Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	cfg := ParseConfig()
	return &Service{store, km, fc, wp, cfg}
}

func (s *Service) Create(c context.Context, sync bool, address string, raw templates.Raw, tType Type) (*jobs.Job, *Transaction, error) {
	t := &Transaction{}

	job, err := s.wp.AddJob(func() (string, error) {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		// Check if the input is a valid address
		address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
		if err != nil {
			return "", err
		}

		id, err := flow_helpers.LatestBlockId(ctx, s.fc)
		if err != nil {
			return "", err
		}

		a, err := s.km.UserAuthorizer(ctx, flow.HexToAddress(address))
		if err != nil {
			return "", err
		}

		var aa []keys.Authorizer

		// Check if we need to add this account as an authorizer
		if strings.Contains(raw.Code, ": AuthAccount") {
			aa = append(aa, a)
		}

		b, err := templates.NewBuilderFromRaw(raw)
		if err != nil {
			return "", err
		}

		proposer := a
		if address == s.cfg.AdminAccountAddress {
			proposer, err = s.km.AdminProposalKey(ctx)
			if err != nil {
				return "", err
			}
		}

		if err := New(t, id, b, tType, proposer, a, aa); err != nil {
			return t.TransactionId, err
		}

		// Send the transaction

		if err := t.Send(ctx, s.fc); err != nil {
			return t.TransactionId, err
		}

		// Insert to datastore
		if err := s.store.InsertTransaction(t); err != nil {
			return t.TransactionId, err
		}

		// Wait for the transaction to be sealed
		if err := t.Wait(ctx, s.fc); err != nil {
			return t.TransactionId, err
		}

		// Update in datastore
		err = s.store.UpdateTransaction(t)

		return t.TransactionId, err
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

	return job, t, err
}

// List returns all transactions in the datastore for a given account.
func (s *Service) List(tType Type, address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return []Transaction{}, err
	}

	o := datastore.ParseListOptions(limit, offset)

	return s.store.Transactions(tType, address, o)
}

// Details returns a specific transaction.
func (s *Service) Details(tType Type, address, transactionId string) (result Transaction, err error) {
	// Check if the input is a valid address
	address, err = flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return
	}

	// Check if the input is a valid transaction id
	if err = flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return
	}

	// Get from datastore
	result, err = s.store.Transaction(tType, address, transactionId)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("transaction not found"),
		}
		return
	}

	return
}

// Execute a script
func (s *Service) ExecuteScript(ctx context.Context, raw templates.Raw) (cadence.Value, error) {
	return s.fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(raw.Code),
		templates.MustDecodeArgs(raw.Arguments),
	)
}

func (s *Service) UpdateTransaction(t *Transaction) error {
	return s.store.UpdateTransaction(t)
}

func (s *Service) GetOrCreateTransaction(transactionId string) *Transaction {
	return s.store.GetOrCreateTransaction(transactionId)
}
