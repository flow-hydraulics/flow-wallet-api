package templates

import (
	"strings"
	"testing"

	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
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

	if cfg1.enabledTokenAddresses == nil {
		t.Fatal("expected there to be enabled tokens")
	}
}

func TestParseName(t *testing.T) {
	n := (&Token{Name: "FUSD"}).ParseName()
	if n != [3]string{"FUSD", "FUSD", "fusd"} {
		t.Error("invalid output for FUSD")
	}

	n = (&Token{Name: "fusd"}).ParseName()
	if n != [3]string{"FUSD", "FUSD", "fusd"} {
		t.Error("invalid output for fusd")
	}

	n = (&Token{Name: "FlowToken"}).ParseName()
	if n != [3]string{"FlowToken", "FLOW_TOKEN", "flowToken"} {
		t.Error("invalid output for FlowToken")
	}

	n = (&Token{Name: "flow-token"}).ParseName()
	if n != [3]string{"FlowToken", "FLOW_TOKEN", "flowToken"} {
		t.Error("invalid output for flow-token")
	}
}

func TestParseGenericFungibleTransfer(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		g := FungibleTransferCode(
			Token{Name: "FlowToken"},
		)

		c := Code(&Template{Source: template_strings.TransferFlow})

		if g != c {
			t.Error("expected outputs to equal")
		}
	})

	t.Run("FlowToken with non-standard addresses", func(t *testing.T) {
		g := FungibleTransferCode(
			Token{Name: "FlowToken", Address: "some_other_tokenaddress"},
		)

		c := Code(&Template{Source: template_strings.TransferFlow})

		if g == c {
			t.Error("expected outputs not to equal")
		}

		if i := strings.Index(g, "import FlowToken from some_other_tokenaddress"); i == -1 {
			t.Error("expected to find non-standard tokenaddress")
		}
	})

	t.Run("FUSD", func(t *testing.T) {
		g := FungibleTransferCode(
			Token{Name: "FUSD"},
		)

		c := Code(&Template{template_strings.TransferFUSD})

		if g != c {
			t.Error("expected outputs to equal")
		}
	})

	t.Run("FUSD with non-standard addresses", func(t *testing.T) {
		g := FungibleTransferCode(
			Token{Name: "FUSD", Address: "some_other_tokenaddress"},
		)

		c := Code(&Template{template_strings.TransferFUSD})

		if g == c {
			t.Error("expected outputs not to equal")
		}

		if i := strings.Index(g, "import FUSD from some_other_tokenaddress"); i == -1 {
			t.Error("expected to find non-standard tokenaddress")
		}
	})
}
