package tokens

import (
	"context"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	flow_templates "github.com/onflow/flow-go-sdk/templates"
)

// DeployTokenContractForAccount is used for testing purposes.
func (s *ServiceImpl) DeployTokenContractForAccount(ctx context.Context, runSync bool, tokenName, address string) error {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return err
	}

	token, err := s.templates.GetTokenByName(tokenName)
	if err != nil {
		return err
	}

	n := token.Name

	tmplStr, err := template_strings.GetByName(n)
	if err != nil {
		return err
	}

	src, err := templates.TokenCode(s.cfg.ChainID, token, tmplStr)
	if err != nil {
		return err
	}

	c := flow_templates.Contract{Name: n, Source: src}

	err = accounts.AddContract(ctx, s.fc, s.km, address, c, s.cfg.TransactionTimeout)
	if err != nil && !strings.Contains(err.Error(), "cannot overwrite existing contract") {
		return err
	}

	return nil
}
