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

func (s *Service) CreateFtWithdrawal(ctx context.Context, sync bool, token templates.Token, sender, recipient, amount string) (*jobs.Job, *transactions.Transaction, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(sender, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	sender = flow_helpers.HexString(sender)

	err = flow_helpers.ValidateAddress(recipient, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	recipient = flow_helpers.HexString(recipient)

	_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return nil, nil, err
	}

	raw := templates.Raw{
		Code: templates.FungibleTransferCode(token, s.cfg.ChainId),
		Arguments: []templates.Argument{
			_amount,
			cadence.NewAddress(flow.HexToAddress(recipient)),
		},
	}

	return s.ts.Create(ctx, sync, sender, raw, transactions.FtWithdrawal)
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

// DeployTokenContractForAccount is mainly used for testing purposes.
func (s *Service) DeployTokenContractForAccount(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
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
