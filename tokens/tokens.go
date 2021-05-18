// Package tokens provides functions for token handling in Flow blockhain.
// https://docs.onflow.org/core-contracts
package tokens

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/onflow/flow-go-sdk"
)

func TemplatePath() string {
	// TODO: figure out a dynamic way to do this?
	return filepath.Join("cadence")
}

func FungibleTokenContractAddress(chainId flow.ChainID) (string, error) {
	switch chainId {
	default:
		return "", fmt.Errorf("FungibleToken address not found for '%s'", chainId)
	case flow.Emulator:
		return "0xee82856bf20e2aa6", nil
	case flow.Testnet:
		return "0x9a0766d93b6608b7", nil
	case flow.Mainnet:
		return "0xf233dcee88fe0abe", nil
	}
}

func FlowTokenContractAddress(chainId flow.ChainID) (string, error) {
	switch chainId {
	default:
		return "", fmt.Errorf("FlowToken address not found for '%s'", chainId)
	case flow.Emulator:
		return "0x0ae53cb6e3f42a79", nil
	case flow.Testnet:
		return "0x7e60df042a9c0868", nil
	case flow.Mainnet:
		return "0x1654653399040a61", nil
	}
}

func ParseFlowTokenTransactionCode(filename string, chainId flow.ChainID) (string, error) {
	p := filepath.Join(TemplatePath(), "transactions", filename)

	t, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}

	b, err := FungibleTokenContractAddress(chainId)
	if err != nil {
		return "", err
	}

	a, err := FlowTokenContractAddress(chainId)
	if err != nil {
		return "", err
	}

	replacer := strings.NewReplacer(
		"<BaseTokenAddress>", b,
		"<TokenAddress>", a,
	)

	return replacer.Replace(string(t)), nil
}

func ParseTransferFlowToken(chainId flow.ChainID) (string, error) {
	return ParseFlowTokenTransactionCode("transfer_flow.cdc", chainId)
}

func ParseGetFlowTokenBalance(chainId flow.ChainID) (string, error) {
	return ParseFlowTokenTransactionCode("get_balance.cdc", chainId)
}
