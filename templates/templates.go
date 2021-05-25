package templates

import (
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

func ParseCode(template string, chainId flow.ChainID) string {
	r := replacers[chainId]
	return r.Replace(template)
}
