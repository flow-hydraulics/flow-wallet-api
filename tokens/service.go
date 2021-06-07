package tokens

import (
	"context"

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
	km  keys.Manager
	fc  *client.Client
	ts  *transactions.Service
	cfg Config
}

func NewService(
	km keys.Manager,
	fc *client.Client,
	ts *transactions.Service) *Service {
	cfg := ParseConfig()
	return &Service{km, fc, ts, cfg}
}

func (s *Service) CreateFtWithdrawal(ctx context.Context, sync bool, tokenName, sender, recipient, amount string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(sender, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	err = flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return nil, nil, err
	}

	raw := templates.Raw{
		Code: templates.FungibleTransferCode(templates.NewToken(tokenName), s.cfg.ChainId),
		Arguments: []templates.Argument{
			_amount,
			cadence.NewAddress(flow.HexToAddress(recipient)),
		},
	}

	return s.ts.Create(ctx, sync, sender, raw, transactions.FtWithdrawal)
}

func (s *Service) SetupFtForAccount(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	raw := templates.Raw{
		Code: templates.FungibleSetupCode(templates.NewToken(tokenName), s.cfg.ChainId),
	}

	return s.ts.Create(ctx, sync, address, raw, transactions.FtSetup)
}

// DeployTokenContractForAccount is mainly used for testing purposes.
func (s *Service) DeployTokenContractForAccount(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(address, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}

	n := (&templates.Token{Name: tokenName}).ParseName()[0]
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
