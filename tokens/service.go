package tokens

import (
	"context"
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/flow_helpers"
	"github.com/eqlabs/flow-wallet-api/jobs"
	"github.com/eqlabs/flow-wallet-api/keys"
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/eqlabs/flow-wallet-api/templates/template_strings"
	"github.com/eqlabs/flow-wallet-api/transactions"
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
	store        Store
	km           keys.Manager
	fc           *client.Client
	transactions *transactions.Service
	templates    *templates.Service
	accounts     *accounts.Service
	cfg          Config
}

func NewService(
	store Store,
	km keys.Manager,
	fc *client.Client,
	txs *transactions.Service,
	tes *templates.Service,
	acs *accounts.Service,
) *Service {
	cfg := ParseConfig()
	return &Service{store, km, fc, txs, tes, acs, cfg}
}

func (s *Service) Setup(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, nil, err
	}

	raw := templates.Raw{
		Code: token.Setup,
	}

	var txType transactions.Type

	switch token.Type {
	case templates.FT:
		txType = transactions.FtSetup
	case templates.NFT:
		txType = transactions.NftSetup
	}

	job, tx, err := s.transactions.Create(ctx, sync, address, raw, txType)

	// Handle adding token to account in database
	go func() {
		if err := job.Wait(true); err != nil && !strings.Contains(err.Error(), "vault exists") {
			return
		}

		// Won't return an error on duplicate keys as it uses FirstOrCreate
		err = s.store.InsertAccountToken(&AccountToken{
			AccountAddress: address,
			TokenAddress:   token.Address,
			TokenName:      token.Name,
			TokenType:      token.Type,
		})
		if err != nil {
			fmt.Printf("error while adding account token: %s\n", err)
		}
	}()

	return job, tx, err
}

func (s *Service) AccountTokens(address string, tType *templates.TokenType) ([]AccountToken, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	return s.store.AccountTokens(address, tType)
}

// Details is used to get the accounts balance (or similar for NFTs) for a token.
func (s *Service) Details(ctx context.Context, tokenName, address string) (*Details, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
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
	case templates.NFT:
		// Continue normal flow
	default:
		return nil, fmt.Errorf("unknown token type")
	}

	r := templates.Raw{
		Code: token.Balance,
		Arguments: []templates.Argument{
			cadence.NewAddress(flow.HexToAddress(address)),
		},
	}

	b, err := s.transactions.ExecuteScript(ctx, r)
	if err != nil {
		return nil, err
	}

	return &Details{TokenName: token.Name, Balance: b}, nil
}

func (s *Service) CreateWithdrawal(ctx context.Context, runSync bool, sender string, request WithdrawalRequest) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the sender is a valid address
	sender, err := flow_helpers.ValidateAddress(sender, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	// Check if the recipient is a valid address
	recipient, err := flow_helpers.ValidateAddress(request.Recipient, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	token, err := s.templates.GetTokenByName(request.TokenName)
	if err != nil {
		return nil, nil, err
	}

	switch token.Type {
	case templates.FT:
		// Continue normal flow
	case templates.NFT:
		return nil, nil, fmt.Errorf("not yet implemented")
	default:
		return nil, nil, fmt.Errorf("unknown token type")
	}

	// Convert amount to a cadence value
	c_amount, err := cadence.NewUFix64(request.FtAmount)
	if err != nil {
		return nil, nil, err
	}

	// Raw transfer template
	raw := templates.Raw{
		Code: token.Transfer,
		Arguments: []templates.Argument{
			c_amount,
			cadence.NewAddress(flow.HexToAddress(recipient)),
		},
	}

	// Create the transaction
	job, tx, err := s.transactions.Create(ctx, runSync, sender, raw, transactions.FtTransfer)

	// Initialise the transfer object
	t := &TokenTransfer{
		RecipientAddress: recipient,
		FtAmount:         request.FtAmount,
		TokenName:        token.Name,
	}

	// Handle database update
	go func() {
		if err := job.Wait(true); err != nil {
			// There was an error regarding the transaction
			// Don't store a token transfer
			// TODO: record the error so it can be delivered to client, job has the error
			// but the client may not have the job id (if using sync)
			return
		}
		t.TransactionId = tx.TransactionId
		if err := s.store.InsertFungibleTokenTransfer(t); err != nil {
			fmt.Printf("error while inserting token transfer: %s\n", err)
		}
	}()

	return job, tx, err
}

func (s *Service) RegisterDeposit(token *templates.Token, transactionId, amount, recipient string) error {
	// Check if the input address is a valid address
	recipient, err := flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return err
	}

	// Check if the transaction id is valid
	if err := flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return err
	}

	switch token.Type {
	case templates.FT:
		// Continue normal flow
	case templates.NFT:
		return fmt.Errorf("not yet implemented")
	default:
		return fmt.Errorf("unknown token type")
	}

	// Get existing transaction or create one
	tx := s.transactions.GetOrCreateTransaction(transactionId)
	if tx.TransactionType == transactions.Unknown {
		// Transaction was just created
		// Deposit mostly likely did not originate in this wallet service
		tx.TransactionType = transactions.FtTransfer
		flowTx, err := s.fc.GetTransaction(context.Background(), flow.HexToID(tx.TransactionId))
		if err != nil {
			return err
		}
		tx.PayerAddress = flow_helpers.FormatAddress(flowTx.Payer)
		if err := s.transactions.UpdateTransaction(tx); err != nil {
			return err
		}
	}

	// TODO: Add AccountToken for account if it doesn't already exist (it should but just to be sure)

	// Check for existing deposit
	if _, err := s.store.FungibleTokenDeposit(recipient, token.Name, tx.TransactionId); err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		// Deposit not found, continue
	} else {
		// Deposit found, we are done
		return nil
	}

	// Create and store a new token transfer
	t := &TokenTransfer{
		TransactionId:    tx.TransactionId,
		RecipientAddress: recipient,
		FtAmount:         amount,
		TokenName:        token.Name,
	}

	return s.store.InsertFungibleTokenTransfer(t)
}

func (s *Service) ListTransfers(transferType, address, tokenName string) ([]*TokenTransfer, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case templates.FT:
		// Continue normal flow
	case templates.NFT:
		return nil, fmt.Errorf("not yet implemented")
	default:
		return nil, fmt.Errorf("unknown token type")
	}

	switch transferType {
	default:
		return nil, fmt.Errorf("unknown transfer type %s", transferType)
	case transferTypeWithdrawal:
		return s.store.FungibleTokenWithdrawals(address, token.Name)
	case transferTypeDeposit:
		return s.store.FungibleTokenDeposits(address, token.Name)
	}
}

func (s *Service) ListWithdrawals(address, tokenName string) ([]*FungibleTokenWithdrawal, error) {
	tt, err := s.ListTransfers(transferTypeWithdrawal, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*FungibleTokenWithdrawal, len(tt))
	for i, t := range tt {
		w := t.Withdrawal()
		res[i] = &w
	}
	return res, nil
}

func (s *Service) ListDeposits(address, tokenName string) ([]*FungibleTokenDeposit, error) {
	tt, err := s.ListTransfers(transferTypeDeposit, address, tokenName)
	if err != nil {
		return nil, err
	}
	res := make([]*FungibleTokenDeposit, len(tt))
	for i, t := range tt {
		d := t.Deposit()
		res[i] = &d
	}
	return res, nil
}

func (s *Service) GetTransfer(transferType, address, tokenName, transactionId string) (*TokenTransfer, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
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
		// Continue normal flow
	case templates.NFT:
		return nil, fmt.Errorf("not yet implemented")
	default:
		return nil, fmt.Errorf("unknown token type")
	}

	switch transferType {
	default:
		return nil, fmt.Errorf("unknown transfer type %s", transferType)
	case transferTypeWithdrawal:
		return s.store.FungibleTokenWithdrawal(address, token.Name, transactionId)
	case transferTypeDeposit:
		return s.store.FungibleTokenDeposit(address, token.Name, transactionId)
	}
}

func (s *Service) GetWithdrawal(address, tokenName, transactionId string) (*FungibleTokenWithdrawal, error) {
	t, err := s.GetTransfer(transferTypeWithdrawal, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	w := t.Withdrawal()
	return &w, nil
}

func (s *Service) GetDeposit(address, tokenName, transactionId string) (*FungibleTokenDeposit, error) {
	t, err := s.GetTransfer(transferTypeDeposit, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	d := t.Deposit()
	return &d, nil
}

// DeployTokenContractForAccount is mainly used for testing purposes.
func (s *Service) DeployTokenContractForAccount(ctx context.Context, runSync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return nil, nil, err
	}

	n := token.Name

	tmplStr, err := template_strings.GetByName(n)
	if err != nil {
		return nil, nil, err
	}

	src := templates.TokenCode(token, tmplStr)

	c := flow_templates.Contract{Name: n, Source: src}

	t, err := accounts.AddContract(ctx, s.fc, s.km, address, c)
	if err != nil {
		return nil, t, err
	}

	return nil, t, nil
}
