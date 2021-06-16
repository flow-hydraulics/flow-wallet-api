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
	"github.com/eqlabs/flow-wallet-service/templates"
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

func (s *Service) Create(c context.Context, sync bool, address string, raw templates.Raw, tType Type) (*jobs.Job, *Transaction, error) {
	t := &Transaction{}

	job, err := s.wp.AddJob(func() (string, error) {
		ctx := c
		if !sync {
			ctx = context.Background()
		}

		// Check if the input is a valid address
		if err := flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
			return "", err
		}
		address = flow_helpers.HexString(address)

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

		if err := New(t, id, b, tType, a, a, aa); err != nil {
			return t.TransactionId, err
		}

		// Send the transaction

		if err := t.Send(ctx, s.fc); err != nil {
			return t.TransactionId, err
		}

		// Insert to datastore
		if err := s.db.InsertTransaction(t); err != nil {
			return t.TransactionId, err
		}

		// Wait for the transaction to be sealed
		if err := t.Wait(ctx, s.fc); err != nil {
			return t.TransactionId, err
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

	return job, t, err
}

// List returns all transactions in the datastore for a given account.
func (s *Service) List(tType Type, address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	if err := flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
		return []Transaction{}, err
	}
	address = flow_helpers.HexString(address)

	o := datastore.ParseListOptions(limit, offset)

	return s.db.Transactions(tType, address, o)
}

// Details returns a specific transaction.
func (s *Service) Details(tType Type, address, transactionId string) (result Transaction, err error) {
	// Check if the input is a valid address
	if err = flow_helpers.ValidateAddress(address, s.cfg.ChainId); err != nil {
		return
	}
	address = flow_helpers.HexString(address)

	// Check if the input is a valid transaction id
	if err = flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return
	}

	// Get from datastore
	result, err = s.db.Transaction(tType, address, transactionId)
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
	value, err := s.fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(raw.Code),
		templates.MustDecodeArgs(raw.Arguments),
	)

	if err != nil {
		return cadence.Void{}, err
	}

	return value, err
}

func (s *Service) UpdateTransaction(t *Transaction) error {
	return s.db.UpdateTransaction(t)
}

func (s *Service) GetOrCreateTransaction(transactionId string) *Transaction {
	return s.db.GetOrCreateTransaction(transactionId)
}
