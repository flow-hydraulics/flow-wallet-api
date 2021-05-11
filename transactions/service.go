package transactions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for transaction HTTP handlers.
type Service struct {
	log *log.Logger
	db  Store
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
}

// NewService initiates a new transaction service.
func NewService(
	l *log.Logger,
	db Store,
	km keys.Manager,
	fc *client.Client,
	wp *jobs.WorkerPool,
) *Service {
	return &Service{l, db, km, fc, wp}
}

func (s *Service) create(ctx context.Context, address string, code string, args []TransactionArg) (*Transaction, error) {
	referenceBlockID, err := flow_helpers.LatestBlockId(ctx, s.fc)
	if err != nil {
		return &EmptyTransaction, err
	}

	authorizer, err := s.km.UserAuthorizer(ctx, address)
	if err != nil {
		return &EmptyTransaction, err
	}

	transaction, err := New(referenceBlockID, code, args, authorizer, authorizer, []keys.Authorizer{})
	if err != nil {
		return &EmptyTransaction, err
	}

	// Send the transaction
	err = transaction.Send(ctx, s.fc)
	if err != nil {
		return transaction, err
	}

	// Set TransactionId
	transaction.TransactionId = transaction.tx.ID().Hex()

	// Insert to datastore
	err = s.db.InsertTransaction(transaction)
	if err != nil {
		return transaction, err
	}

	// Wait for the transaction to be sealed
	err = transaction.Wait(ctx, s.fc)
	if err != nil {
		return transaction, err
	}

	// Update in datastore
	err = s.db.UpdateTransaction(transaction)

	return transaction, err
}

func (s *Service) CreateSync(ctx context.Context, code string, args []TransactionArg, address string) (*Transaction, error) {
	var result *Transaction
	var jobErr error
	var createErr error
	var done bool = false

	_, jobErr = s.wp.AddJob(func() (string, error) {
		result, createErr = s.create(context.Background(), address, code, args)
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

func (s *Service) CreateAsync(code string, args []TransactionArg, address string) (*jobs.Job, error) {
	job, err := s.wp.AddJob(func() (string, error) {
		tx, err := s.create(context.Background(), address, code, args)
		if err != nil {
			s.log.Println(err)
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
func (s *Service) List(address string) ([]Transaction, error) {
	return s.db.Transactions(address)
}

// Details returns a specific transaction.
func (s *Service) Details(address, transactionId string) (result Transaction, err error) {
	// TODO: validate address and transactionId

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
