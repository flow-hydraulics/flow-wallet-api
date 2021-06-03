package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/iancoleman/strcase"
	"github.com/onflow/flow-go-sdk"
)

type chainReplacers map[flow.ChainID]*strings.Replacer
type chainAddresses map[flow.ChainID]string
type templateVariables map[string]chainAddresses

var chains []flow.ChainID = []flow.ChainID{
	flow.Emulator,
	flow.Testnet,
	flow.Mainnet,
}

var replacers chainReplacers

func makeChainReplacers(t templateVariables) chainReplacers {
	r := make(chainReplacers, len(chains))
	for _, c := range chains {
		vv := make([]string, len(t)*2)
		for varname, addressInChain := range t {
			vv = append(vv, varname, addressInChain[c])
		}
		r[c] = strings.NewReplacer(vv...)
	}
	return r
}

func parseName(name string) [3]string {
	// TODO: how to handle these kind of cases?
	if strcase.ToScreamingSnake(name) == "FUSD" {
		return [3]string{
			"FUSD", "FUSD", "fusd",
		}
	}

	return [3]string{
		strcase.ToCamel(name),
		strcase.ToScreamingSnake(name),
		strcase.ToLowerCamel(name),
	}
}

func Code(template string, chainId flow.ChainID) string {
	r := replacers[chainId]
	return r.Replace(template)
}

func GenericFungibleTemplateCode(t, tokenName string, chainId flow.ChainID, addresses ...string) string {
	p := parseName(tokenName)
	camel := p[0]
	snake := p[1]
	lower := p[2]

	r := strings.NewReplacer(
		"TokenName", camel,
		"TOKEN_NAME", snake,
		"tokenName", lower,
	)

	t = r.Replace(t)

	// Replace token address
	if len(addresses) >= 1 && addresses[0] != "" {
		r := strings.NewReplacer(
			fmt.Sprintf("%s_ADDRESS", snake), addresses[0],
		)
		t = r.Replace(t)
	}

	// Replace fungible token contract address
	if len(addresses) >= 2 && addresses[1] != "" {
		r := strings.NewReplacer(
			"FUNGIBLE_TOKEN_ADDRESS", addresses[1],
		)
		t = r.Replace(t)
	}

	return Code(t, chainId)
}

func GenericFungibleTransferCode(tokenName string, chainId flow.ChainID, addresses ...string) string {
	return GenericFungibleTemplateCode(
		template_strings.GenericFungibleTransfer,
		tokenName,
		chainId,
		addresses...,
	)
}

func GenericFungibleSetupCode(tokenName string, chainId flow.ChainID, addresses ...string) string {
	return GenericFungibleTemplateCode(
		template_strings.GenericFungibleSetup,
		tokenName,
		chainId,
		addresses...,
	)
}
