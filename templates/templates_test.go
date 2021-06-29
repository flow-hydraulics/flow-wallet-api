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

func TestParseGenericFungibleTransfer(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		token := &Token{Name: "FlowToken", Address: "test-address", NameLowerCase: "flowToken"}
		c := FungibleTransferCode(token)
		if !strings.Contains(c, fmt.Sprintf("import FlowToken from %s", token.Address)) {
			t.Error("expected to find import statement for token address")
		}
	})

	t.Run("FUSD", func(t *testing.T) {
		token := &Token{Name: "FUSD", Address: "test-address", NameLowerCase: "fusd"}
		c := FungibleTransferCode(token)
		if !strings.Contains(c, fmt.Sprintf("import FUSD from %s", token.Address)) {
			t.Error("expected to find import statement for token address")
		}
	})

}
