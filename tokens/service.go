package tokens

import (
	"context"
	"fmt"
	"sync"

	"github.com/eqlabs/flow-wallet-service/accounts"
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	flow_templates "github.com/onflow/flow-go-sdk/templates"
)

const (
	transferTypeWithdrawal = "withdrawal"
	transferTypeDeposit    = "deposit"
)

type Service struct {
	db  Store
	km  keys.Manager
	fc  *client.Client
	ts  *transactions.Service
	cfg Config
}

func NewService(
	db Store,
	km keys.Manager,
	fc *client.Client,
	ts *transactions.Service) *Service {
	cfg := ParseConfig()
	return &Service{db, km, fc, ts, cfg}
}

func (s *Service) List() []templates.Token {
	return templates.EnabledTokens()
}

func (s *Service) Details(ctx context.Context, tokenName, address string) (TokenDetails, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return TokenDetails{}, err
	}
	address = flow_helpers.HexString(address)

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return TokenDetails{}, err
	}

	r := templates.Raw{
		Code: templates.FungibleBalanceCode(token),
		Arguments: []templates.Argument{
			cadence.NewAddress(flow.HexToAddress(address)),
		},
	}

	b, err := s.ts.ExecuteScript(ctx, r)
	if err != nil {
		return TokenDetails{}, err
	}

	return TokenDetails{Name: token.CanonName(), Balance: b.String()}, nil
}

func (s *Service) CreateFtWithdrawal(ctx context.Context, runSync bool, tokenName, sender, recipient, amount string) (*jobs.Job, *FungibleTokenTransfer, error) {
	// Check if the sender is a valid address
	err := flow_helpers.ValidateAddress(sender, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	// Make sure correct format is used
	sender = flow_helpers.HexString(sender)

	// Check if the recipient is a valid address
	err = flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	// Make sure correct format is used
	recipient = flow_helpers.HexString(recipient)

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return nil, nil, err
	}

	// Convert amount to a cadence value
	c_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return nil, nil, err
	}

	// Raw transfer template
	raw := templates.Raw{
		Code: templates.FungibleTransferCode(token),
		Arguments: []templates.Argument{
			c_amount,
			cadence.NewAddress(flow.HexToAddress(recipient)),
		},
	}

	// Create the transaction
	job, tx, err := s.ts.Create(ctx, runSync, sender, raw, transactions.FtTransfer)

	// Initialise the transfer object
	t := &FungibleTokenTransfer{
		RecipientAddress: recipient,
		Amount:           amount,
		TokenName:        token.CanonName(),
	}

	// Handle database update
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		job.Wait(true) // Ignore the error
		t.TransactionId = tx.TransactionId
		s.db.InsertFungibleTokenTransfer(t) // TODO: handle error
	}()

	if runSync {
		wg.Wait()
	}

	return job, t, err
}

func (s *Service) RegisterFtDeposit(transactionId string, tokenName, amount, recipient string) error {
	// Check if the recipient is a valid address
	err := flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return err
	}
	// Make sure correct format is used
	recipient = flow_helpers.HexString(recipient)

	// Check if the transaction id is valid
	err = flow_helpers.ValidateTransactionId(transactionId)
	if err != nil {
		return err
	}

	// Get existing transaction or create one
	tx := s.ts.GetOrCreateTransaction(transactionId)
	if tx.TransactionType == transactions.Unknown {
		// Transaction was just created
		// Deposit mostly likely did not originate in this wallet service
		tx.TransactionType = transactions.FtTransfer
		flowTx, err := s.fc.GetTransaction(context.Background(), flow.HexToID(tx.TransactionId))
		if err != nil {
			return err
		}
		tx.PayerAddress = flow_helpers.FormatAddress(flowTx.Payer)
		err = s.ts.UpdateTransaction(tx)
		if err != nil {
			return err
		}
	}

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return err
	}

	// Check for existing deposit, ignore error
	t, _ := s.db.FungibleTokenDeposit(recipient, token.CanonName(), tx.TransactionId)
	if t != nil {
		// Found an existing deposit, return
		return nil
	}

	// Create and store a new token transfer
	t = &FungibleTokenTransfer{
		TransactionId:    tx.TransactionId,
		RecipientAddress: recipient,
		Amount:           amount,
		TokenName:        token.CanonName(),
	}

	return s.db.InsertFungibleTokenTransfer(t)
}

func (s *Service) ListFtTransfers(transferType, address, tokenName string) ([]*FungibleTokenTransfer, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}
	address = flow_helpers.HexString(address)

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return nil, err
	}

	switch transferType {
	default:
		return nil, fmt.Errorf("unknown transfer type %s", transferType)
	case transferTypeWithdrawal:
		return s.db.FungibleTokenWithdrawals(address, token.CanonName())
	case transferTypeDeposit:
		return s.db.FungibleTokenDeposits(address, token.CanonName())
	}
}

func (s *Service) ListFtWithdrawals(address, tokenName string) ([]*FungibleTokenWithdrawal, error) {
	tt, err := s.ListFtTransfers(transferTypeWithdrawal, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*FungibleTokenWithdrawal, len(tt))
	for i, t := range tt {
		res[i] = t.Withdrawal()
	}
	return res, nil
}

func (s *Service) ListFtDeposits(address, tokenName string) ([]*FungibleTokenDeposit, error) {
	tt, err := s.ListFtTransfers(transferTypeDeposit, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*FungibleTokenDeposit, len(tt))
	for i, t := range tt {
		res[i] = t.Deposit()
	}
	return res, nil
}

func (s *Service) GetFtTransfer(transferType, address, tokenName, transactionId string) (*FungibleTokenTransfer, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}
	address = flow_helpers.HexString(address)

	// Check if the input is a valid transaction id
	err = flow_helpers.ValidateTransactionId(transactionId)
	if err != nil {
		return nil, err
	}

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return nil, err
	}

	switch transferType {
	default:
		return nil, fmt.Errorf("unknown transfer type %s", transferType)
	case transferTypeWithdrawal:
		return s.db.FungibleTokenWithdrawal(address, token.CanonName(), transactionId)
	case transferTypeDeposit:
		return s.db.FungibleTokenDeposit(address, token.CanonName(), transactionId)
	}
}

func (s *Service) GetFtWithdrawal(address, tokenName, transactionId string) (*FungibleTokenWithdrawal, error) {
	t, err := s.GetFtTransfer(transferTypeWithdrawal, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	return t.Withdrawal(), nil
}

func (s *Service) GetFtDeposit(address, tokenName, transactionId string) (*FungibleTokenDeposit, error) {
	t, err := s.GetFtTransfer(transferTypeDeposit, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	return t.Deposit(), nil
}

// DeployTokenContractForAccount is mainly used for testing purposes.
func (s *Service) DeployTokenContractForAccount(ctx context.Context, runSync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	address = flow_helpers.HexString(address)

	token, err := templates.NewToken(tokenName)
	if err != nil {
		return nil, nil, err
	}

	n := token.CanonName()

	t_str, err := template_strings.GetByName(n)
	if err != nil {
		return nil, nil, err
	}

	src := templates.Code(&templates.Template{Source: t_str})

	c := flow_templates.Contract{Name: n, Source: src}

	t, err := accounts.AddContract(ctx, s.fc, s.km, address, c)
	if err != nil {
		return nil, t, err
	}

	return nil, t, nil
}
