package template_strings

import (
	"fmt"
	"strings"
	"testing"
)

var tokens = BatchedFungibleOpsInfo{
	FungibleTokenContractAddress: "0xFungibleTokenContractAddress",
	Tokens: []FungibleTokenInfo{
		{
			ContractName:       "TokenA",
			Address:            "0x1",
			VaultStoragePath:   "TokenA.VaultStoragePath",
			ReceiverPublicPath: "TokenA.VaultReceiverPubPath",
			BalancePublicPath:  "TokenA.VaultBalancePubPath",
		},
		{
			ContractName:       "TokenB",
			Address:            "0x2",
			VaultStoragePath:   "/storage/tokenBVault",
			ReceiverPublicPath: "/public/tokenBReceiver",
			BalancePublicPath:  "/public/tokenBBalance",
		},
	},
}

func TestAccountCreation(t *testing.T) {
	result, err := CreateAccountAndSetupTransaction(tokens)
	if err != nil {
		t.Error(err)
	}

	checkStrings := []string{
		"import FungibleToken from 0xFungibleTokenContractAddress",
		"import TokenA from 0x1",
		"import TokenB from 0x2",
		"account.save(<-TokenA.createEmptyVault(), to: TokenA.VaultStoragePath)",
		"target: TokenA.VaultStoragePath",
		"account.save(<-TokenB.createEmptyVault(), to: /storage/tokenBVault)",
		"target: /storage/tokenBVault",
	}

	ok, failedCheck := containsAll(result, checkStrings)
	if !ok {
		fmt.Println(result)
		t.Errorf("result doesn't contain: %s", failedCheck)
	}
}

func TestAddFungibleTokens(t *testing.T) {
	result, err := AddFungibleTokenVaultBatchTransaction(tokens)
	if err != nil {
		t.Error(err)
	}

	checkStrings := []string{
		"import FungibleToken from 0xFungibleTokenContractAddress",
		"import TokenA from 0x1",
		"import TokenB from 0x2",
		"account.save(<-TokenA.createEmptyVault(), to: TokenA.VaultStoragePath)",
		"target: TokenA.VaultStoragePath",
		"account.save(<-TokenB.createEmptyVault(), to: /storage/tokenBVault)",
		"target: /storage/tokenBVault",
		"if account.borrow<&TokenA.Vault>(from: TokenA.VaultStoragePath) == nil {",
		"if account.borrow<&TokenB.Vault>(from: /storage/tokenBVault) == nil {",
	}

	ok, failedCheck := containsAll(result, checkStrings)
	if !ok {
		fmt.Println(result)
		t.Errorf("result doesn't contain: %s", failedCheck)
	}
}

func containsAll(result string, checks []string) (bool, string) {
	for _, check := range checks {
		if !strings.Contains(result, check) {
			return false, check
		}
	}
	return true, ""
}
