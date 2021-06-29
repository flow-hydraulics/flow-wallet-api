package templates

import (
	"fmt"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
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

func TestTokenFromEvent(t *testing.T) {
	e := flow.Event{
		Type: "A.0ae53cb6e3f42a79.FlowToken.TokensDeposited",
	}
	token, err := TokenFromEvent(e, flow.Emulator)
	if err != nil {
		t.Fatal(err)
	}
	if token.Address != "0x0ae53cb6e3f42a79" {
		t.Error("invalid token address")
	}
	if token.Name != "FlowToken" {
		t.Error("invalid token name")
	}
}

func TestParseGenericFungibleTransfer(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		token, _ := NewToken("FlowToken")
		c := FungibleTransferCode(token)
		if !strings.Contains(c, fmt.Sprintf("import FlowToken from %s", token.Address)) {
			t.Error("expected to find correct token address")
		}
	})

	t.Run("FlowToken with non-standard addresses", func(t *testing.T) {
		token := &Token{Name: "FlowToken", Address: "some_other_tokenaddress", NameLowerCase: "flowToken"}
		c := FungibleTransferCode(token)
		if !strings.Contains(c, "import FlowToken from some_other_tokenaddress") {
			t.Error("expected to find non-standard token address")
		}
	})

	t.Run("FUSD", func(t *testing.T) {
		token, _ := NewToken("FUSD")
		c := FungibleTransferCode(token)
		if !strings.Contains(c, fmt.Sprintf("import FUSD from %s", token.Address)) {
			t.Error("expected to find correct token address")
		}
	})

	t.Run("FUSD with non-standard addresses", func(t *testing.T) {
		token := &Token{Name: "FUSD", Address: "some_other_tokenaddress", NameLowerCase: "fusd"}
		c := FungibleTransferCode(token)
		if !strings.Contains(c, "import FUSD from some_other_tokenaddress") {
			t.Error("expected to find non-standard token address")
		}
	})
}
