// Package tokens provides functions for token handling in Flow blockhain.
// https://docs.onflow.org/core-contracts
package tokens

import (
	"path/filepath"
	"strings"

	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/onflow/flow-go-sdk"
)

func TemplatePath() string {
	// TODO: figure out a dynamic way to do this?
	return filepath.Join("cadence")
}

func ParseCode(template string, chainId flow.ChainID) (string, error) {
	a1, err := FungibleTokenContractAddress(chainId)
	if err != nil {
		return "", err
	}

	a2, err := FlowTokenContractAddress(chainId)
	if err != nil {
		return "", err
	}

	r := strings.NewReplacer(
		"FUNGIBLE_TOKEN_ADDRESS", a1,
		"FLOW_TOKEN_ADDRESS", a2,
	)

	return r.Replace(template), nil
}

func ParseTransferFlowToken(chainId flow.ChainID) (string, error) {
	t := templates.TransferFlow
	return ParseCode(t, chainId)
}

func ParseGetFlowTokenBalance(chainId flow.ChainID) (string, error) {
	t := templates.GetFlowBalance
	return ParseCode(t, chainId)
}
