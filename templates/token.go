package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/iancoleman/strcase"
	"github.com/onflow/flow-go-sdk"
)

type Token struct {
	Name string
}

func NewToken(name string) Token {
	return Token{Name: name}
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

func fungibleTemplateCode(tmpl_str string, token Token, chainId flow.ChainID, addresses ...string) string {
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
	if len(addresses) >= 1 && addresses[0] != "" {
		r := strings.NewReplacer(
			fmt.Sprintf("%s_ADDRESS", snake), addresses[0],
		)
		tmpl_str = r.Replace(tmpl_str)
	}

	// Replace fungible token contract address
	if len(addresses) >= 2 && addresses[1] != "" {
		r := strings.NewReplacer(
			"FUNGIBLE_TOKEN_ADDRESS", addresses[1],
		)
		tmpl_str = r.Replace(tmpl_str)
	}

	return Code(&Template{Source: tmpl_str}, chainId)
}

func FungibleTransferCode(token Token, chainId flow.ChainID, addresses ...string) string {
	return fungibleTemplateCode(
		template_strings.GenericFungibleTransfer,
		token,
		chainId,
		addresses...,
	)
}

func FungibleSetupCode(token Token, chainId flow.ChainID, addresses ...string) string {
	return fungibleTemplateCode(
		template_strings.GenericFungibleSetup,
		token,
		chainId,
		addresses...,
	)
}
