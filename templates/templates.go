package templates

import (
	"strings"

	"github.com/onflow/flow-go-sdk"
)

type Template struct {
	Source string
}

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

// Code converts a Template to string (source code).
func Code(t *Template) string {
	c := parseConfig().ChainId
	return replacers[c].Replace(t.Source)
}
