package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/iancoleman/strcase"
	"github.com/onflow/flow-go-sdk"
)

type Token struct {
	Name    string `json:"tokenName"`
	Address string `json:"tokenAddress"`
}

func NewToken(name, address string) Token {
	return Token{Name: name, Address: address}
}

func (t *Token) ParseName() [3]string {
	// TODO: how to handle these kind of cases?
	if strcase.ToScreamingSnake(t.Name) == "FUSD" {
		return [3]string{
			"FUSD", "FUSD", "fusd",
		}
	}

	return [3]string{
		strcase.ToCamel(t.Name),
		strcase.ToScreamingSnake(t.Name),
		strcase.ToLowerCamel(t.Name),
	}
}

func fungibleTemplateCode(tmpl_str string, token Token, chainId flow.ChainID) string {
	p := token.ParseName()
	camel := p[0]
	snake := p[1]
	lower := p[2]

	r := strings.NewReplacer(
		"TokenName", camel,
		"TOKEN_NAME", snake,
		"tokenName", lower,
	)

	tmpl_str = r.Replace(tmpl_str)

	// Replace token address
	if token.Address != "" {
		r := strings.NewReplacer(
			fmt.Sprintf("%s_ADDRESS", snake), token.Address,
		)
		tmpl_str = r.Replace(tmpl_str)
	}

	return Code(&Template{Source: tmpl_str}, chainId)
}

func FungibleTransferCode(token Token, chainId flow.ChainID) string {
	return fungibleTemplateCode(
		template_strings.GenericFungibleTransfer,
		token,
		chainId,
	)
}

func FungibleSetupCode(token Token, chainId flow.ChainID) string {
	return fungibleTemplateCode(
		template_strings.GenericFungibleSetup,
		token,
		chainId,
	)
}

func FungibleBalanceCode(token Token, chainId flow.ChainID) string {
	return fungibleTemplateCode(
		template_strings.GenericFungibleBalance,
		token,
		chainId,
	)
}
