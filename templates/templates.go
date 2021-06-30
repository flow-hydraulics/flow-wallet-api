package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/onflow/flow-go-sdk"
)

type Template struct {
	Source string
}

type Token struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name" gorm:"uniqueIndex;not null"` // Declaration name
	NameLowerCase string    `json:"nameLowerCase,omitempty"`          // For generic fungible token transaction templates
	Address       string    `json:"address" gorm:"not null"`
	Setup         string    `json:"setup,omitempty"`    // Setup cadence code
	Transfer      string    `json:"transfer,omitempty"` // Transfer cadence code
	Balance       string    `json:"balance,omitempty"`  // Balance cadence code
	Type          TokenType `json:"type"`
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

// Code converts a Template to string (source code).
func Code(t *Template) string {
	c := parseConfig().ChainId
	return replacers[c].Replace(t.Source)
}

func FungibleTransferCode(token *Token) string {
	return tokenCode(
		template_strings.GenericFungibleTransfer,
		token,
	)
}

func FungibleSetupCode(token *Token) string {
	return tokenCode(
		template_strings.GenericFungibleSetup,
		token,
	)
}

func FungibleBalanceCode(token *Token) string {
	return tokenCode(
		template_strings.GenericFungibleBalance,
		token,
	)
}

func tokenCode(tmplStr string, token *Token) string {
	r := strings.NewReplacer(
		"TOKEN_DECLARATION_NAME", token.Name,
		"TOKEN_ADDRESS", token.Address,
		"TOKEN_VAULT", fmt.Sprintf("%sVault", token.NameLowerCase),
		"TOKEN_RECEIVER", fmt.Sprintf("%sReceiver", token.NameLowerCase),
		"TOKEN_BALANCE", fmt.Sprintf("%sBalance", token.NameLowerCase),
	)

	return Code(&Template{Source: r.Replace(tmplStr)})
}

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
