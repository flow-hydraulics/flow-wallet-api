package transactions

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
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for transaction HTTP handlers.
type Service struct {
	db  Store
	Km  keys.Manager
	Fc  *client.Client
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

func (s *Service) Create(c context.Context, sync bool, address, code string, args []Argument, tType Type) (*jobs.Job, *Transaction, error) {
	var transaction *Transaction

	job, err := s.wp.AddJob(func() (string, error) {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		// Check if the input is a valid address
		err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
		if err != nil {
			return "", err
		}

		id, err := flow_helpers.LatestBlockId(ctx, s.Fc)
		if err != nil {
			return "", err
		}

		a, err := s.Km.UserAuthorizer(ctx, flow.HexToAddress(address))
		if err != nil {
			return "", err
		}

		var aa []keys.Authorizer

		// Check if we need to add this account as an authorizer
		if strings.Contains(code, ": AuthAccount") {
			aa = append(aa, a)
		}

		t, err := New(id, code, args, tType, a, a, aa)
		if err != nil {
			return "", err
		}

		transaction = t

		// Send the transaction
		err = t.Send(ctx, s.Fc)
		if err != nil {
			return "", err
		}

		// Insert to datastore
		err = s.db.InsertTransaction(t)
		if err != nil {
			return "", err
		}

		// Wait for the transaction to be sealed
		err = t.Wait(ctx, s.Fc)
		if err != nil {
			return "", err
		}

		// Update in datastore
		err = s.db.UpdateTransaction(t)

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

	return job, transaction, err
}

// List returns all transactions in the datastore for a given account.
func (s *Service) List(address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return []Transaction{}, err
	}

	o := datastore.ParseListOptions(limit, offset)

	return s.db.Transactions(flow_helpers.HexString(address), o)
}

// Details returns a specific transaction.
func (s *Service) Details(address, transactionId string) (result Transaction, err error) {
	// Check if the input is a valid address
	err = flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return
	}

	// Check if the input is a valid transaction id
	err = flow_helpers.ValidateTransactionId(transactionId)
	if err != nil {
		return
	}

	// Get from datastore
	result, err = s.db.Transaction(flow_helpers.HexString(address), transactionId)
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
	value, err := s.Fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(code),
		MustDecodeArgs(args),
	)

	if err != nil {
		return cadence.Void{}, err
	}

	return value, err
}
