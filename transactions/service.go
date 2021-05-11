package transactions

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for transaction HTTP handlers.
type Service struct {
	log *log.Logger
	db  TransactionStore
	km  keys.Manager
	fc  *client.Client
	wp  *jobs.WorkerPool
}

// NewService initiates a new transaction service.
func NewService(
	l *log.Logger,
	db TransactionStore,
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

	// Send the transaction
	transaction.Send(ctx, s.fc)

	// Set TransactionId
	transaction.TransactionId = transaction.tx.ID().Hex()

	// Insert to datastore
	s.db.InsertTransaction(transaction)

	// Wait for the transaction to be sealed
	transaction.Wait(ctx, s.fc)

	// Update in datastore
	s.db.UpdateTransaction(transaction)

	return transaction, err
}

func (s *Service) CreateSync(ctx context.Context, code string, args []TransactionArg, address string) (*Transaction, error) {
	return s.create(ctx, address, code, args)
}

func (s *Service) CreateAsync(ctx context.Context, code string, args []TransactionArg, address string) (*jobs.Job, error) {
	job, err := s.wp.AddJob(func() (string, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tx, err := s.create(ctx, address, code, args)
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
			Err:        fmt.Errorf("job not found"),
		}
		return
	}

	return
}
