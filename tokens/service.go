package tokens

import (
	"context"
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

func (s *Service) Details(ctx context.Context, token templates.Token, address string) (TokenDetails, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return TokenDetails{}, err
	}
	address = flow_helpers.HexString(address)

	r := templates.Raw{
		Code: templates.FungibleBalanceCode(token, s.cfg.ChainId),
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

func (s *Service) CreateFtWithdrawal(ctx context.Context, runSync bool, token templates.Token, sender, recipient, amount string) (*jobs.Job, *FungibleTokenTransfer, error) {
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

	// Convert amount to a cadence value
	c_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return nil, nil, err
	}

	// Raw transfer template
	raw := templates.Raw{
		Code: templates.FungibleTransferCode(token, s.cfg.ChainId),
		Arguments: []templates.Argument{
			c_amount,
			cadence.NewAddress(flow.HexToAddress(recipient)),
		},
	}

	// Create the transaction
	job, _, err := s.ts.Create(ctx, runSync, sender, raw, transactions.FtWithdrawal)

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
		t.TransactionId = job.Result
		s.db.InsertFungibleTokenTransfer(t) // TODO: handle error
	}()

	if runSync {
		wg.Wait()
	}

	return job, t, err
}

func (s *Service) ListFtWithdrawals(address string, token templates.Token) ([]FungibleTokenTransfer, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, err
	}
	address = flow_helpers.HexString(address)

	return s.db.FungibleTokenWithdrawals(address, token.CanonName())
}

func (s *Service) GetFtWithdrawal(address string, token templates.Token, transactionId string) (FungibleTokenTransfer, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return FungibleTokenTransfer{}, err
	}
	address = flow_helpers.HexString(address)

	// Check if the input is a valid transaction id
	err = flow_helpers.ValidateTransactionId(transactionId)
	if err != nil {
		return FungibleTokenTransfer{}, err
	}

	return s.db.FungibleTokenWithdrawal(address, token.CanonName(), transactionId)
}

// DeployTokenContractForAccount is mainly used for testing purposes.
func (s *Service) DeployTokenContractForAccount(ctx context.Context, runSync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	address = flow_helpers.HexString(address)

	n := (&templates.Token{Name: tokenName}).CanonName()
	t_str, err := template_strings.GetByName(n)
	if err != nil {
		return nil, nil, err
	}

	src := templates.Code(&templates.Template{Source: t_str}, s.cfg.ChainId)

	c := flow_templates.Contract{Name: n, Source: src}

	t, err := accounts.AddContract(ctx, s.fc, s.km, address, c)
	if err != nil {
		return nil, t, err
	}

	return nil, t, nil
}
