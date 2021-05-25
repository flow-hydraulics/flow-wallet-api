package tokens

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

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
