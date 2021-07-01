package templates

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestConfig(t *testing.T) {
	if err := godotenv.Load("../.env.test"); err != nil {
		t.Fatal(err)
	}

	cfg1 := parseConfig()
	cfg2 := parseConfig()

	if cfg1 != cfg2 {
		t.Fatal("expected configs to point to the same address")
	}

	if cfg1.enabledTokens == nil {
		t.Fatal("expected there to be enabled tokens")
	}
}

func TestParsing(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		token := &Token{Name: "FlowToken", Address: "test-address", NameLowerCase: "flowToken"}
		c := FungibleTransferCode(token)
		if strings.Contains(c, ".cdc") {
			t.Error("exptected all cadence file references to have been replaced")
		}
		if !strings.Contains(c, fmt.Sprintf("import FlowToken from %s", token.Address)) {
			t.Error("expected to find import statement for token address")
		}
	})

	t.Run("FUSD", func(t *testing.T) {
		token := &Token{Name: "FUSD", Address: "test-address", NameLowerCase: "fusd"}
		c := FungibleTransferCode(token)
		if strings.Contains(c, ".cdc") {
			t.Error("exptected all cadence file references to have been replaced")
		}
		if !strings.Contains(c, fmt.Sprintf("import FUSD from %s", token.Address)) {
			t.Error("expected to find import statement for token address")
		}
	})

	t.Run("ExampleNFT", func(t *testing.T) {
		token := &Token{Name: "ExampleNFT", Address: "test-address"}
		c := TokenCode(
			token,
			`
				import NonFungibleToken from "../contracts/NonFungibleToken.cdc"
				import ExampleNFT from "../contracts/ExampleNFT.cdc"

				transaction(recipient: Address, withdrawID: UInt64) {
						prepare(signer: AuthAccount) {
								// get the recipients public account object
								let recipient = getAccount(recipient)

								// borrow a reference to the signer's NFT collection
								let collectionRef = signer.borrow<&ExampleNFT.Collection>(from: ExampleNFT.CollectionStoragePath) ?? panic("Could not borrow a reference to the owner's collection")

								// borrow a public reference to the receivers collection
								let depositRef = recipient.getCapability(ExampleNFT.CollectionPublicPath)!.borrow<&{NonFungibleToken.CollectionPublic}>()!

								// withdraw the NFT from the owner's collection
								let nft <- collectionRef.withdraw(withdrawID: withdrawID)

								// Deposit the NFT in the recipient's collection
								depositRef.deposit(token: <-nft)
						}
				}
			`)
		if strings.Contains(c, ".cdc") {
			t.Error("exptected all cadence file references to have been replaced")
		}
		if !strings.Contains(c, fmt.Sprintf("import ExampleNFT from %s", token.Address)) {
			t.Error("expected to find import statement for token address")
		}
	})
}
