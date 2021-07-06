package tokens

import (
	"context"
	"fmt"
	"strconv"
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
	queryTypeWithdrawal = "withdrawal"
	queryTypeDeposit    = "deposit"
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
		return nil, fmt.Errorf("unsupported token type: %s", token.Type)
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

	var txType transactions.Type
	var arguments []templates.Argument = make([]templates.Argument, 2)

	switch token.Type {
	case templates.FT:
		txType = transactions.FtTransfer
		amount, err := cadence.NewUFix64(request.FtAmount)
		if err != nil {
			return nil, nil, err
		}
		arguments[0] = amount
		arguments[1] = cadence.NewAddress(flow.HexToAddress(recipient))
	case templates.NFT:
		txType = transactions.NftTransfer
		arguments[0] = cadence.NewAddress(flow.HexToAddress(recipient))
		arguments[1] = cadence.NewUInt64(request.NftID)
	default:
		return nil, nil, fmt.Errorf("unsupported token type: %s", token.Type)
	}

	// Raw transfer template
	raw := templates.Raw{
		Code:      token.Transfer,
		Arguments: arguments,
	}

	// Create the transaction
	job, tx, err := s.transactions.Create(ctx, runSync, sender, raw, txType)

	// Initialise the transfer object
	t := &TokenTransfer{
		RecipientAddress: recipient,
		FtAmount:         request.FtAmount,
		NftID:            request.NftID,
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
		if err := s.store.InsertTokenTransfer(t); err != nil {
			fmt.Printf("error while inserting token transfer: %s\n", err)
		}
	}()

	return job, tx, err
}

func (s *Service) ListTransfers(queryType, address, tokenName string) ([]*TokenTransfer, error) {
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

func (s *Service) ListWithdrawals(address, tokenName string) ([]*TokenWithdrawal, error) {
	tt, err := s.ListTransfers(queryTypeWithdrawal, address, tokenName)
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

func (s *Service) ListDeposits(address, tokenName string) ([]*TokenDeposit, error) {
	tt, err := s.ListTransfers(queryTypeDeposit, address, tokenName)
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

func (s *Service) GetTransfer(queryType, address, tokenName, transactionId string) (*TokenTransfer, error) {
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

func (s *Service) GetWithdrawal(address, tokenName, transactionId string) (*TokenWithdrawal, error) {
	t, err := s.GetTransfer(queryTypeWithdrawal, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	w := t.Withdrawal()
	return &w, nil
}

func (s *Service) GetDeposit(address, tokenName, transactionId string) (*TokenDeposit, error) {
	t, err := s.GetTransfer(queryTypeDeposit, address, tokenName, transactionId)
	if err != nil {
		return nil, err
	}
	d := t.Deposit()
	return &d, nil
}

func (s *Service) RegisterDeposit(token *templates.Token, transactionId, amountOrNftID, recipient string) error {
	var ftAmount string
	var nftId uint64

	switch token.Type {
	case templates.FT:
		ftAmount = amountOrNftID
	case templates.NFT:
		u, err := strconv.ParseUint(amountOrNftID, 10, 64)
		if err != nil {
			return err
		}
		nftId = u
	default:
		return fmt.Errorf("unsupported token type: %s", token.Type)
	}

	// Check if the input address is a valid address
	recipient, err := flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return err
	}

	// Check if the transaction id is valid
	if err := flow_helpers.ValidateTransactionId(transactionId); err != nil {
		return err
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
	if _, err := s.store.TokenDeposit(recipient, tx.TransactionId, token); err != nil {
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
		FtAmount:         ftAmount,
		NftID:            nftId,
		TokenName:        token.Name,
	}

	return s.store.InsertTokenTransfer(t)
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
