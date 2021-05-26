package templates

import (
	"fmt"
	"strings"

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

func ParseCode(chainId flow.ChainID, template string) string {
	r := replacers[chainId]
	return r.Replace(template)
}

func ParseGenericFungibleTransfer(
	chainId flow.ChainID,
	TokenName, TOKEN_NAME, tokenName,
	baseAddress, tokenAddress string,
) string {
	r := strings.NewReplacer(
		"TokenName", TokenName,
		"TOKEN_NAME", TOKEN_NAME,
		"tokenName", tokenName,
	)

	t := r.Replace(GenericFungibleTransfer)

	if baseAddress != "" {
		r := strings.NewReplacer(
			"FUNGIBLE_TOKEN_ADDRESS", baseAddress,
		)
		t = r.Replace(t)
	}

	if tokenAddress != "" {
		r := strings.NewReplacer(
			fmt.Sprintf("%s_ADDRESS", TOKEN_NAME), tokenAddress,
		)
		t = r.Replace(t)
	}

	return ParseCode(chainId, t)
}
