package template_strings

import (
	"fmt"
	"testing"
)

var tokens = BatchedFungibleOpsInfo{
	FungibleTokenAddress: "0xFUN",
	Tokens: []FungibleTokenInfo{
		{
			ContractName:       "FiatToken",
			Address:            "0x123",
			VaultStoragePath:   "FiatToken.VaultStoragePath",
			ReceiverPublicPath: "FiatToken.VaultReceiverPubPath",
			BalancePublicPath:  "FiatToken.VaultBalancePubPath",
		},
		{
			ContractName:       "FiatToken",
			Address:            "0x123",
			VaultStoragePath:   "FiatToken.VaultStoragePath",
			ReceiverPublicPath: "FiatToken.VaultReceiverPubPath",
			BalancePublicPath:  "FiatToken.VaultBalancePubPath",
		},
	},
}

func TestFungibleToken(t *testing.T) {
	result, err := CreateAccountAndSetupTransaction(tokens)
	fmt.Println(result)
	if err == nil {
		t.Error(err)
	}

	// TODO test result
}

func TestAddFungibleToken(t *testing.T) {
	result, err := AddFungibleTokenVaultBatchTransaction(tokens)
	fmt.Println(result)
	if err == nil {
		t.Error(err)
	}

	// TODO test result
}
