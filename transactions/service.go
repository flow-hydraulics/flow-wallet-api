package transactions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/eqlabs/flow-wallet-service/datastore"
	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for transaction HTTP handlers.
type Service struct {
	db  Store
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
	cfg Config
}

// NewService initiates a new transaction service.
func NewService(
	db Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	cfg := ParseConfig()
	return &Service{db, km, fc, wp, cfg}
}

func (s *Service) create(ctx context.Context, address flow.Address, code string, args []Argument) (*Transaction, error) {
	id, err := flow_helpers.LatestBlockId(ctx, s.fc)
	if err != nil {
		return &EmptyTransaction, err
	}

	a, err := s.km.UserAuthorizer(ctx, address)
	if err != nil {
		return &EmptyTransaction, err
	}

	var aa []keys.Authorizer

	// Check if we need to add this account as an authorizer
	if strings.Contains(code, ": AuthAccount") {
		aa = append(aa, a)
	}

	t, err := New(id, code, args, a, a, aa)
	if err != nil {
		return &EmptyTransaction, err
	}

	// Send the transaction
	err = t.Send(ctx, s.fc)
	if err != nil {
		return t, err
	}

	// Insert to datastore
	err = s.db.InsertTransaction(t)
	if err != nil {
		return t, err
	}

	// Wait for the transaction to be sealed
	err = t.Wait(ctx, s.fc)
	if err != nil {
		return t, err
	}

	// Update in datastore
	err = s.db.UpdateTransaction(t)

	return t, err
}

func (s *Service) CreateSync(ctx context.Context, code string, args []Argument, address string) (*Transaction, error) {
	var result *Transaction
	var jobErr error
	var createErr error
	var done bool = false

	// Check if the input is a valid address
	err := accounts.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	_, jobErr = s.wp.AddJob(func() (string, error) {
		result, createErr = s.create(context.Background(), flow.HexToAddress(address), code, args)
		done = true
		if createErr != nil {
			return "", createErr
		}
		return result.TransactionId, nil
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

func (s *Service) CreateAsync(code string, args []Argument, address string) (*jobs.Job, error) {
	// Check if the input is a valid address
	err := accounts.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	job, err := s.wp.AddJob(func() (string, error) {
		tx, err := s.create(context.Background(), flow.HexToAddress(address), code, args)
		if err != nil {
			return "", err
		}
		return tx.TransactionId, nil
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

// List returns all transactions in the datastore for a given account.
func (s *Service) List(address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	err := accounts.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return []Transaction{}, err
	}

	o := datastore.ParseListOptions(limit, offset)

	return s.db.Transactions(address, o)
}

// Details returns a specific transaction.
func (s *Service) Details(address, transactionId string) (result Transaction, err error) {
	// Check if the input is a valid address
	err = accounts.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return
	}

	// Check if the input is a valid transaction id
	err = ValidateTransactionId(transactionId)
	if err != nil {
		return
	}

	// Get from datastore
	result, err = s.db.Transaction(address, transactionId)
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
func (s *Service) ExecuteScript(ctx context.Context, code string, args []Argument) (cadence.Value, error) {
	value, err := s.fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(code),
		MustDecodeArgs(args),
	)

	if err != nil {
		return cadence.Void{}, err
	}

	return value, err
}
