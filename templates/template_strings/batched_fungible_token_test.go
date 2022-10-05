package template_strings

import (
	"fmt"
	"testing"
)

var tokens = BatchedFungibleOpsInfo{
	FungibleTokenContractAddress: "0xFUN",
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

func TestAccountCreation(t *testing.T) {
	result, err := CreateAccountAndSetupTransaction(tokens)
	fmt.Println(result)
	if err != nil {
		t.Error(err)
	}

	// TODO test result
}

func TestAddFungibleTokens(t *testing.T) {
	result, err := AddFungibleTokenVaultBatchTransaction(tokens)
	fmt.Println(result)
	if err != nil {
		t.Error(err)
	}

	// TODO test result
}
