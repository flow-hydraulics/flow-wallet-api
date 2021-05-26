package templates

import (
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
)

func TestParseGenericFungibleTransfer(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		g := ParseGenericFungibleTransfer(
			flow.Emulator,
			"FlowToken", "FLOW_TOKEN", "flowToken",
			"", "",
		)

		c := ParseCode(flow.Emulator, TransferFlow)

		if g != c {
			t.Error("expected outputs to equal")
		}
	})

	t.Run("FlowToken with non-standard addresses", func(t *testing.T) {
		g := ParseGenericFungibleTransfer(
			flow.Emulator,
			"FlowToken", "FLOW_TOKEN", "flowToken",
			"some_other_baseaddress", "some_other_tokenaddress",
		)

		c := ParseCode(flow.Emulator, TransferFlow)

		if g == c {
			t.Error("expected outputs not to equal")
		}

		if i := strings.Index(g, "import FungibleToken from some_other_baseaddress"); i == -1 {
			t.Error("expected to find non-standard baseaddress")
		}

		if i := strings.Index(g, "import FlowToken from some_other_tokenaddress"); i == -1 {
			t.Error("expected to find non-standard tokenaddress")
		}
	})

	t.Run("FUSD", func(t *testing.T) {
		g := ParseGenericFungibleTransfer(
			flow.Emulator,
			"FUSD", "FUSD", "fusd",
			"", "",
		)

		c := ParseCode(flow.Emulator, TransferFUSD)

		if g != c {
			t.Error("expected outputs to equal")
		}
	})

	t.Run("FUSD with non-standard addresses", func(t *testing.T) {
		g := ParseGenericFungibleTransfer(
			flow.Emulator,
			"FUSD", "FUSD", "fusd",
			"some_other_baseaddress", "some_other_tokenaddress",
		)

		c := ParseCode(flow.Emulator, TransferFlow)

		if g == c {
			t.Error("expected outputs not to equal")
		}

		if i := strings.Index(g, "import FungibleToken from some_other_baseaddress"); i == -1 {
			t.Error("expected to find non-standard baseaddress")
		}

		if i := strings.Index(g, "import FUSD from some_other_tokenaddress"); i == -1 {
			t.Error("expected to find non-standard tokenaddress")
		}
	})
}
