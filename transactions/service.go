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
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
)

type Service interface {
	Create(ctx context.Context, sync bool, proposerAddress string, code string, args []Argument, tType Type) (*jobs.Job, *Transaction, error)
	Sign(ctx context.Context, proposerAddress string, code string, args []Argument) (*SignedTransaction, error)
	List(limit, offset int) ([]Transaction, error)
	ListForAccount(tType Type, address string, limit, offset int) ([]Transaction, error)
	Details(ctx context.Context, transactionId string) (*Transaction, error)
	DetailsForAccount(ctx context.Context, tType Type, address, transactionId string) (*Transaction, error)
	ExecuteScript(ctx context.Context, code string, args []Argument) (cadence.Value, error)
	UpdateTransaction(t *Transaction) error
	GetOrCreateTransaction(transactionId string) *Transaction
}

// ServiceImpl defines the API for transaction HTTP handlers.
type ServiceImpl struct {
	store         Store
	km            keys.Manager
	fc            flow_helpers.FlowClient
	wp            jobs.WorkerPool
	cfg           *configs.Config
	txRateLimiter ratelimit.Limiter
}

// NewService initiates a new transaction service.
func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc flow_helpers.FlowClient,
	wp jobs.WorkerPool,
	opts ...ServiceOption,
) Service {
	var defaultTxRatelimiter = ratelimit.NewUnlimited()

	// TODO(latenssi): safeguard against nil config?
	svc := &ServiceImpl{store, km, fc, wp, cfg, defaultTxRatelimiter}

	for _, opt := range opts {
		opt(svc)
	}

	if wp == nil {
		panic("workerpool nil")
	}

	// Register asynchronous job executor.
	wp.RegisterExecutor(TransactionJobType, svc.executeTransactionJob)

	return svc
}

func (s *ServiceImpl) Create(ctx context.Context, sync bool, proposerAddress string, code string, args []Argument, tType Type) (*jobs.Job, *Transaction, error) {
	transaction, err := s.newTransaction(ctx, proposerAddress, code, args, tType)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting new transaction: %w", err)
	}

	if err := s.store.InsertTransaction(transaction); err != nil {
		return nil, nil, fmt.Errorf("error while inserting transaction in db: %w", err)
	}

	if !sync {
		// Async
		job, err := s.wp.CreateJob(TransactionJobType, transaction.TransactionId)
		if err != nil {
			return nil, nil, fmt.Errorf("error while creating job: %w", err)
		}

		if err := s.wp.Schedule(job); err != nil {
			return nil, nil, fmt.Errorf("error while scheduling job: %w", err)
		}

		return job, transaction, nil

	} else {
		// Sync
		if err := s.sendTransaction(ctx, transaction); err != nil {
			return nil, nil, err
		}

		return nil, transaction, nil
	}
}

func (s *ServiceImpl) Sign(ctx context.Context, proposerAddress string, code string, args []Argument) (*SignedTransaction, error) {
	flowTx, err := s.buildFlowTransaction(ctx, proposerAddress, code, args)
	if err != nil {
		return nil, err
	}

	return &SignedTransaction{Transaction: *flowTx}, nil
}

// List returns all transactions in the datastore.
func (s *ServiceImpl) List(limit, offset int) ([]Transaction, error) {
	o := datastore.ParseListOptions(limit, offset)
	return s.store.Transactions(o)
}

// ListForAccount returns all transactions in the datastore for a given account.
func (s *ServiceImpl) ListForAccount(tType Type, address string, limit, offset int) ([]Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return []Transaction{}, err
	}

	o := datastore.ParseListOptions(limit, offset)

	return s.store.TransactionsForAccount(tType, address, o)
}

// Details returns a specific transaction.
func (s *ServiceImpl) Details(ctx context.Context, transactionId string) (*Transaction, error) {
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
func (s *ServiceImpl) DetailsForAccount(ctx context.Context, tType Type, address, transactionId string) (*Transaction, error) {
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
func (s *ServiceImpl) ExecuteScript(ctx context.Context, code string, args []Argument) (cadence.Value, error) {
	return s.fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(code),
		MustDecodeArgs(args),
	)
}

func (s *ServiceImpl) UpdateTransaction(t *Transaction) error {
	return s.store.UpdateTransaction(t)
}

func (s *ServiceImpl) GetOrCreateTransaction(transactionId string) *Transaction {
	return s.store.GetOrCreateTransaction(transactionId)
}

func (s *ServiceImpl) buildFlowTransaction(ctx context.Context, proposerAddress, code string, arguments []Argument) (*flow.Transaction, error) {
	latestBlockID, err := flow_helpers.LatestBlockId(ctx, s.fc)
	if err != nil {
		return nil, err
	}

	// Admin should always be the payer of the transaction fees.
	payer, err := s.km.AdminAuthorizer(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while getting admin authorizer for payer: %w", err)
	}

	proposer, err := s.getProposalAuthorizer(ctx, proposerAddress)
	if err != nil {
		return nil, err
	}

	flowTx := flow.NewTransaction()
	flowTx.
		SetReferenceBlockID(*latestBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address).
		SetGasLimit(maxGasLimit).
		SetScript([]byte(code))

	for _, arg := range arguments {
		cv, err := ArgAsCadence(arg)
		if err != nil {
			return nil, err
		}

		err = flowTx.AddArgument(cv)
		if err != nil {
			return nil, err
		}
	}

	// Add authorizers. We assume proposer is always the sole authorizer
	// https://github.com/flow-hydraulics/flow-wallet-api/issues/79
	flowTx.AddAuthorizer(proposer.Address)

	// Proposer signs the payload (unless proposer == payer).
	if !proposer.Equals(payer) {
		if err := flowTx.SignPayload(proposer.Address, proposer.Key.Index, proposer.Signer); err != nil {
			return nil, err
		}
	}

	// Payer signs the envelope
	if err := flowTx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer); err != nil {
		return nil, err
	}

	return flowTx, nil
}

func (s *ServiceImpl) newTransaction(ctx context.Context, proposerAddress string, code string, args []Argument, tType Type) (*Transaction, error) {
	tx := &Transaction{
		ProposerAddress: proposerAddress,
		TransactionType: tType,
	}

	flowTx, err := s.buildFlowTransaction(ctx, proposerAddress, code, args)
	if err != nil {
		return nil, fmt.Errorf("error while building transaction: %w", err)
	}

	tx.TransactionId = flowTx.ID().Hex()
	tx.FlowTransaction = flowTx.Encode()

	return tx, nil
}

func (s *ServiceImpl) getProposalAuthorizer(ctx context.Context, proposerAddress string) (keys.Authorizer, error) {
	// Validate the input address.
	proposerAddress, err := flow_helpers.ValidateAddress(proposerAddress, s.cfg.ChainID)
	if err != nil {
		return keys.Authorizer{}, err
	}

	var proposer keys.Authorizer
	if proposerAddress == s.cfg.AdminAddress {
		proposer, err = s.km.AdminProposalKey(ctx)
		if err != nil {
			return keys.Authorizer{}, fmt.Errorf("error while getting admin authorizer: %w", err)
		}
	} else {
		proposer, err = s.km.UserAuthorizer(ctx, flow.HexToAddress(proposerAddress))
		if err != nil {
			return keys.Authorizer{}, fmt.Errorf("error while getting user authorizer: %w", err)
		}
	}

	return proposer, nil
}

func (s *ServiceImpl) sendTransaction(ctx context.Context, tx *Transaction) error {
	// TODO: we should "recreate" the transaction as proposal key sequence numbering
	// might have gotten out of sync by now (in async situations)

	flowTx, err := flow.DecodeTransaction(tx.FlowTransaction)
	if err != nil {
		return err
	}

	// Check if transaction has been sent already.
	_, err = s.fc.GetTransaction(ctx, flowTx.ID())
	if err != nil {
		rpcErr, ok := err.(grpc.RPCError)
		if !ok {
			// The error wasn't from gRPC.
			return err
		}

		if rpcErr.GRPCStatus().Code() != codes.NotFound {
			// Something unexpected went wrong in the gRPC call or in the Access API.
			return err
		}

		// The Flow transaction was not found. All good. Continue.
	}

	// Ratelimit
	s.txRateLimiter.Take()

	resp, err := flow_helpers.SendAndWait(ctx, s.fc, *flowTx, s.cfg.TransactionTimeout)
	if err != nil {
		return err
	}

	tx.Events = resp.Events

	return nil
}
