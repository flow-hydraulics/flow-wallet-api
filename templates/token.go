package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/onflow/flow-go-sdk"
)

type Token struct {
	Name    string `json:"name"` // Declaration name
	Address string `json:"address"`
	lcName  string `json:"-"` // lowerCamelCase name for generic fungible token transaction templates
}

func EnabledTokens() map[string]Token {
	return parseConfig().enabledTokens
}

func NewToken(name string) (*Token, error) {
	// Assume the name is in declaration name format
	token, ok := EnabledTokens()[name]
	if !ok {
		return nil, fmt.Errorf("token %s not enabled, make sure to use the declaration name format", name)
	}
	return &token, nil
}

func TokenFromEvent(e flow.Event, chainId flow.ChainID) (*Token, error) {
	// Example event:
	// A.0ae53cb6e3f42a79.FlowToken.TokensDeposited
	ss := strings.Split(e.Type, ".")

	eAddress, err := flow_helpers.ValidateAddress(ss[1], chainId)
	if err != nil {
		return nil, err
	}

	t, err := NewToken(ss[2])
	if err != nil {
		return nil, err
	}

	tAddress, err := flow_helpers.ValidateAddress(t.Address, chainId)
	if err != nil {
		return nil, err
	}

	if eAddress != tAddress {
		return nil, fmt.Errorf("addresses do not match for %s, from event %s, from config %s", t.Name, eAddress, tAddress)
	}

	return t, nil
}

func (t *Token) IsEnabled() bool {
	for _, e := range EnabledTokens() {
		if t.Name == e.Name && t.Address == e.Address {
			return true
		}
	}
	return false
}

func tokenTemplateCode(tmpl_str string, token *Token) string {
	r := strings.NewReplacer(
		"TOKEN_DECLARATION_NAME", token.Name,
		"TOKEN_ADDRESS", token.Address,
		"TOKEN_VAULT", fmt.Sprintf("%sVault", token.lcName),
		"TOKEN_RECEIVER", fmt.Sprintf("%sReceiver", token.lcName),
		"TOKEN_BALANCE", fmt.Sprintf("%sBalance", token.lcName),
	)

	tmpl_str = r.Replace(tmpl_str)

	return Code(&Template{Source: tmpl_str})
}

func FungibleTransferCode(token *Token) string {
	return tokenTemplateCode(
		template_strings.GenericFungibleTransfer,
		token,
	)
}

func FungibleSetupCode(token *Token) string {
	return tokenTemplateCode(
		template_strings.GenericFungibleSetup,
		token,
	)
}

func FungibleBalanceCode(token *Token) string {
	return tokenTemplateCode(
		template_strings.GenericFungibleBalance,
		token,
	)
}
