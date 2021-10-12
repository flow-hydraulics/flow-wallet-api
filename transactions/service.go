package transactions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
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
	cfg   *configs.Config
}

// NewService initiates a new transaction service.
func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	// TODO(latenssi): safeguard against nil config?
	return &Service{store, km, fc, wp, cfg}
}

func (s *Service) Create(c context.Context, sync bool, proposerAddress string, raw templates.Raw, tType Type) (*jobs.Job, *Transaction, error) {
	transaction := &Transaction{}

	process := func(jobResult *jobs.Result) error {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		// Check if the input is a valid address
		proposerAddress, err := flow_helpers.ValidateAddress(proposerAddress, s.cfg.ChainID)
		if err != nil {
			return err
		}

		latestBlockId, err := flow_helpers.LatestBlockId(ctx, s.fc)
		if err != nil {
			return err
		}

		var (
			payer    keys.Authorizer
			proposer keys.Authorizer
		)

		// Admin should always be the payer of the transaction fees
		payer, err = s.km.AdminAuthorizer(ctx)
		if err != nil {
			return err
		}

		if proposerAddress == s.cfg.AdminAddress {
			proposer, err = s.km.AdminProposalKey(ctx)
			if err != nil {
				return err
			}
		} else {
			proposer, err = s.km.UserAuthorizer(ctx, flow.HexToAddress(proposerAddress))
			if err != nil {
				return err
			}

		}

		// We assume proposer is always the sole authorizer
		// https://github.com/flow-hydraulics/flow-wallet-api/issues/79
		authorizers := []keys.Authorizer{proposer}

		builder, err := templates.NewBuilderFromRaw(raw)
		if err != nil {
			return err
		}

		// Init a new transaction
		err = New(transaction, *latestBlockId, builder, tType, proposer, payer, authorizers)
		jobResult.TransactionID = transaction.TransactionId // Update job result
		if err != nil {
			return err
		}

		// Insert to datastore
		if err := s.store.InsertTransaction(transaction); err != nil {
			return err
		}

		// Send and wait for the transaction to be sealed
		result, err := flow_helpers.SendAndWait(ctx, s.fc, *builder.Tx, s.cfg.TransactionTimeout)
		if result != nil {
			// Record for possible JSON response
			transaction.Events = result.Events
		}
		if err != nil {
			return err
		}

		// Update in datastore
		if err := s.store.UpdateTransaction(transaction); err != nil {
			return err
		}

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

	return job, transaction, err
}

func (s *Service) Sign(c context.Context, proposerAddress string, raw templates.Raw, tType Type) (*SignedTransaction, error) {
	transaction := &Transaction{}
	var signedTransaction *SignedTransaction

	process := func(jobResult *jobs.Result) error {
		ctx := c

		// Check if the input is a valid address.
		proposerAddress, err := flow_helpers.ValidateAddress(proposerAddress, s.cfg.ChainID)
		if err != nil {
			return err
		}

		latestBlockId, err := flow_helpers.LatestBlockId(ctx, s.fc)
		if err != nil {
			return err
		}

		var (
			payer    keys.Authorizer
			proposer keys.Authorizer
		)

		// Admin should always be the payer of the transaction fees.
		payer, err = s.km.AdminAuthorizer(ctx)
		if err != nil {
			return err
		}

		if flow_helpers.HexString(proposerAddress) == flow_helpers.HexString(s.cfg.AdminAddress) {
			proposer, err = s.km.AdminProposalKey(ctx)
			if err != nil {
				return err
			}
		} else {
			proposer, err = s.km.UserAuthorizer(ctx, flow.HexToAddress(proposerAddress))
			if err != nil {
				return err
			}

		}

		// We assume proposer is always the sole authorizer
		// https://github.com/flow-hydraulics/flow-wallet-api/issues/79
		authorizers := []keys.Authorizer{proposer}

		builder, err := templates.NewBuilderFromRaw(raw)
		if err != nil {
			return err
		}

		// Init a new transaction
		err = New(transaction, *latestBlockId, builder, tType, proposer, payer, authorizers)
		jobResult.TransactionID = transaction.TransactionId // Update job result
		if err != nil {
			return err
		}

		signedTransaction = &SignedTransaction{Transaction: *builder.Tx}
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
		return nil, err
	}

	// Signing is always synchronous, but executes through worker pool.
	err = job.Wait(true)

	return signedTransaction, err
}

// List returns all transactions in the datastore.
func (s *Service) List(limit, offset int) ([]Transaction, error) {
	o := datastore.ParseListOptions(limit, offset)
	return s.store.Transactions(o)
}

// ListForAccount returns all transactions in the datastore for a given account.
func (s *Service) ListForAccount(tType Type, address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return []Transaction{}, err
	}

	o := datastore.ParseListOptions(limit, offset)

	return s.store.TransactionsForAccount(tType, address, o)
}

// Details returns a specific transaction.
func (s *Service) Details(ctx context.Context, transactionId string) (*Transaction, error) {
	// Check if the input is a valid transaction id
	if err := flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return nil, err
	}

	// Get from datastore
	transaction, err := s.store.Transaction(transactionId)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("transaction not found"),
		}
		return nil, err
	}

	result, err := s.fc.GetTransactionResult(ctx, flow.HexToID(transactionId))
	if err != nil {
		return nil, err
	}

	transaction.Events = result.Events

	return &transaction, nil
}

// DetailsForAccount returns a specific transaction.
func (s *Service) DetailsForAccount(ctx context.Context, tType Type, address, transactionId string) (*Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	// Check if the input is a valid transaction id
	if err = flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return nil, err
	}

	// Get from datastore
	transaction, err := s.store.TransactionForAccount(tType, address, transactionId)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("transaction not found"),
		}
		return nil, err
	}

	result, err := s.fc.GetTransactionResult(ctx, flow.HexToID(transactionId))
	if err != nil {
		return nil, err
	}

	transaction.Events = result.Events

	return &transaction, nil
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
