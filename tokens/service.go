package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

const (
	queryTypeWithdrawal = "withdrawal"
	queryTypeDeposit    = "deposit"
)

type Service interface {
	Setup(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error)
	AddAccountToken(tokenName, address string) error
	AccountTokens(address string, tType templates.TokenType) ([]AccountToken, error)
	Details(ctx context.Context, tokenName, address string) (*Details, error)
	CreateWithdrawal(ctx context.Context, sync bool, sender string, request WithdrawalRequest) (*jobs.Job, *transactions.Transaction, error)
	ListWithdrawals(address, tokenName string) ([]*TokenWithdrawal, error)
	ListDeposits(address, tokenName string) ([]*TokenDeposit, error)
	GetWithdrawal(address, tokenName, transactionId string) (*TokenWithdrawal, error)
	GetDeposit(address, tokenName, transactionId string) (*TokenDeposit, error)
	RegisterDeposit(ctx context.Context, token *templates.Token, transactionId flow.Identifier, recipient accounts.Account, amountOrNftID string) error

	// DeployTokenContractForAccount is only used in tests
	DeployTokenContractForAccount(ctx context.Context, runSync bool, tokenName, address string) error
}

type ServiceImpl struct {
	store        Store
	km           keys.Manager
	fc           flow_helpers.FlowClient
	wp           jobs.WorkerPool
	transactions transactions.Service
	templates    templates.Service
	accounts     accounts.Service
	cfg          *configs.Config
}

func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc flow_helpers.FlowClient,
	wp jobs.WorkerPool,
	txs transactions.Service,
	tes templates.Service,
	acs accounts.Service,
) Service {
	// TODO(latenssi): safeguard against nil config?

	svc := &ServiceImpl{store, km, fc, wp, txs, tes, acs, cfg}

	if wp == nil {
		panic("workerpool nil")
	}

	// Register asynchronous job executor.
	wp.RegisterExecutor(WithdrawalCreateJobType, svc.executeCreateWithdrawalJob)

	return svc
}

func (s *ServiceImpl) Setup(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, nil, err
	}

	// Check if the account with given address exists
	_, err = s.accounts.Details(address)
	if err != nil {
		return nil, nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, nil, err
	}

	var txType transactions.Type

	switch token.Type {
	case templates.FT:
		txType = transactions.FtSetup
	case templates.NFT:
		txType = transactions.NftSetup
	}

	job, tx, err := s.transactions.Create(ctx, sync, address, token.Setup, nil, txType)

	if err == nil || strings.Contains(err.Error(), "vault exists") {
		// Handle adding token to account in database
		if err := s.store.InsertAccountToken(&AccountToken{
			AccountAddress: address,
			TokenAddress:   token.Address,
			TokenName:      token.Name,
			TokenType:      token.Type,
		}); err != nil {
			log.
				WithFields(log.Fields{"error": err}).
				Warn("Error while adding account token")
		}
	}

	return job, tx, err
}

func (s *ServiceImpl) AddAccountToken(tokenName, address string) error {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return err
	}

	if err := s.store.InsertAccountToken(&AccountToken{
		AccountAddress: address,
		TokenAddress:   token.Address,
		TokenName:      token.Name,
		TokenType:      token.Type,
	}); err != nil {
		return err
	}

	return nil
}

func (s *ServiceImpl) AccountTokens(address string, tType templates.TokenType) ([]AccountToken, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	return s.store.AccountTokens(address, tType)
}

// Details is used to get the accounts balance (or similar for NFTs) for a token.
func (s *ServiceImpl) Details(ctx context.Context, tokenName, address string) (*Details, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	// Get the correct token from database
	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case templates.FT:
		fallthrough
	case templates.NFT:
		// Continue normal flow
	default:
		return nil, fmt.Errorf("unsupported token type: %s", token.Type)
	}

	res, err := s.transactions.ExecuteScript(ctx, token.Balance, []transactions.Argument{cadence.NewAddress(flow.HexToAddress(address))})
	if err != nil {
		return nil, err
	}

	return &Details{TokenName: token.Name, Balance: &Balance{CadenceValue: res}}, nil
}

func (s *ServiceImpl) CreateWithdrawal(ctx context.Context, sync bool, sender string, request WithdrawalRequest) (*jobs.Job, *transactions.Transaction, error) {
	log.WithFields(log.Fields{"sync": sync}).Trace("Create withdrawal")

	if !sync {
		// Async
		attrs := withdrawalCreateJobAttributes{sender, request}
		attrBytes, err := json.Marshal(attrs)
		if err != nil {
			return nil, nil, err
		}

		job, err := s.wp.CreateJob(WithdrawalCreateJobType, "", jobs.WithAttributes(attrBytes))
		if err != nil {
			return nil, nil, err
		}

		err = s.wp.Schedule(job)
		if err != nil {
			return nil, nil, err
		}

		return job, nil, err

	} else {
		// Sync
		transaction, err := s.createWithdrawal(ctx, sender, request)
		if err != nil {
			return nil, nil, err
		}

		return nil, transaction, nil
	}
}

func (s *ServiceImpl) listTransfers(queryType, address, tokenName string) ([]*TokenTransfer, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case templates.FT:
	case templates.NFT:
		// Continue normal flow
	default:
		return nil, fmt.Errorf("unsupported token type: %s", token.Type)
	}

	switch queryType {
	default:
		return nil, fmt.Errorf("unknown transfer type %s", queryType)
	case queryTypeWithdrawal:
		return s.store.TokenWithdrawals(address, token)
	case queryTypeDeposit:
		return s.store.TokenDeposits(address, token)
	}
}

func (s *ServiceImpl) ListWithdrawals(address, tokenName string) ([]*TokenWithdrawal, error) {
	tt, err := s.listTransfers(queryTypeWithdrawal, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*TokenWithdrawal, len(tt))
	for i, t := range tt {
		w := t.Withdrawal()
		res[i] = &w
	}
	return res, nil
}

func (s *ServiceImpl) ListDeposits(address, tokenName string) ([]*TokenDeposit, error) {
	tt, err := s.listTransfers(queryTypeDeposit, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*TokenDeposit, len(tt))
	for i, t := range tt {
		d := t.Deposit()
		res[i] = &d
	}
	return res, nil
}

func (s *ServiceImpl) getTransfer(queryType, address, tokenName, transactionId string) (*TokenTransfer, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	// Check if the input is a valid transaction id
	if err := flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case templates.FT:
		fallthrough
	case templates.NFT:
		// Continue normal flow
	default:
		return nil, fmt.Errorf("unsupported token type: %s", token.Type)
	}

	switch queryType {
	default:
		return nil, fmt.Errorf("unknown query %s", queryType)
	case queryTypeWithdrawal:
		return s.store.TokenWithdrawal(address, transactionId, token)
	case queryTypeDeposit:
		return s.store.TokenDeposit(address, transactionId, token)
	}
}

func (s *ServiceImpl) GetWithdrawal(address, tokenName, transactionId string) (*TokenWithdrawal, error) {
	t, err := s.getTransfer(queryTypeWithdrawal, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	w := t.Withdrawal()
	return &w, nil
}

func (s *ServiceImpl) GetDeposit(address, tokenName, transactionId string) (*TokenDeposit, error) {
	t, err := s.getTransfer(queryTypeDeposit, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	d := t.Deposit()
	return &d, nil
}

// RegisterDeposit is an internal API for registering token deposits from on-chain events.
func (s *ServiceImpl) RegisterDeposit(ctx context.Context, token *templates.Token, transactionId flow.Identifier, recipient accounts.Account, amountOrNftID string) error {
	var (
		ftAmount string
		nftId    uint64
	)

	switch token.Type {
	case templates.FT:
		ftAmount = amountOrNftID
	case templates.NFT:
		var err error
		nftId, err = strconv.ParseUint(amountOrNftID, 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported token type: %s", token.Type)
	}

	// TODO (latenssi): db lock for transaction; could it also allow "syncing" when running multiple instances?

	// Get existing transaction or create one
	transaction := s.transactions.GetOrCreateTransaction(transactionId.Hex())
	flowTx, err := s.fc.GetTransaction(ctx, transactionId)
	if err != nil {
		return err
	}

	if transaction.TransactionType == transactions.Unknown {
		// Transaction was just created
		// Transfer most likely did not originate in this wallet service
		transaction.TransactionType = transactions.FtTransfer
		transaction.ProposerAddress = flow_helpers.FormatAddress(flowTx.ProposalKey.Address)
		if err := s.transactions.UpdateTransaction(transaction); err != nil {
			return err
		}
	}

	// Make sure the token is enabled in the database for the recipient account
	// We are registering a deposit event, so the token must be setup already for the recipient
	err = s.store.InsertAccountToken(&AccountToken{
		AccountAddress: recipient.Address,
		TokenAddress:   token.Address,
		TokenName:      token.Name,
		TokenType:      token.Type,
	})
	if err != nil {
		return err
	}

	// Check for existing deposit
	if _, err := s.store.TokenDeposit(recipient.Address, transaction.TransactionId, token); err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			// Error did not contain "record not found"
			return err
		}
		// Error contains "record not found", proceed
	} else {
		// err == nil, existing deposit found, we are done
		return nil
	}

	// Create and store a new token transfer
	transfer := &TokenTransfer{
		TransactionId:    transaction.TransactionId,
		RecipientAddress: recipient.Address,
		SenderAddress:    flow_helpers.FormatAddress(flowTx.Authorizers[0]),
		FtAmount:         ftAmount,
		NftID:            nftId,
		TokenName:        token.Name,
	}

	if err := s.store.InsertTokenTransfer(transfer); err != nil {
		return err
	}

	return nil
}

// createWithdrawal will synchronously create a withdrawal and store the transfer.
// Used in job execution and sync API calls.
func (s *ServiceImpl) createWithdrawal(ctx context.Context, sender string, request WithdrawalRequest) (*transactions.Transaction, error) {
	// Check if the sender is a valid address
	sender, err := flow_helpers.ValidateAddress(sender, s.cfg.ChainID)
	if err != nil {
		return nil, fmt.Errorf("createWithdrawal sender validation error: %w", err)
	}

	// Check if the recipient is a valid address
	recipient, err := flow_helpers.ValidateAddress(request.Recipient, s.cfg.ChainID)
	if err != nil {
		return nil, fmt.Errorf("createWithdrawal recipient validation error: %w", err)
	}

	token, err := s.templates.GetTokenByName(request.TokenName)
	if err != nil {
		return nil, fmt.Errorf("createWithdrawal could not find token error: %w", err)
	}

	var txType transactions.Type
	var arguments []transactions.Argument = make([]transactions.Argument, 2)

	switch token.Type {
	case templates.FT:
		txType = transactions.FtTransfer
		amount, err := cadence.NewUFix64(request.FtAmount)
		if err != nil {
			return nil, err
		}
		arguments[0] = amount
		arguments[1] = cadence.NewAddress(flow.HexToAddress(recipient))
	case templates.NFT:
		txType = transactions.NftTransfer
		arguments[0] = cadence.NewAddress(flow.HexToAddress(recipient))
		arguments[1] = cadence.NewUInt64(request.NftID)
	default:
		return nil, fmt.Errorf("createWithdrawal unsupported token type: %s", token.Type)
	}

	// Create the transaction, must be sync here
	_, transaction, err := s.transactions.Create(ctx, true, sender, token.Transfer, arguments, txType)
	if err != nil {
		return nil, err
	}

	// Store Transfer in database
	transfer := &TokenTransfer{
		TransactionId:    transaction.TransactionId,
		RecipientAddress: recipient,
		SenderAddress:    sender,
		FtAmount:         request.FtAmount,
		NftID:            request.NftID,
		TokenName:        token.Name,
	}

	if err := s.store.InsertTokenTransfer(transfer); err != nil {
		return nil, err
	}

	return transaction, nil
}
