package templates

import (
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
)

func TestParseName(t *testing.T) {
	n := parseName("FUSD")
	if n != [3]string{"FUSD", "FUSD", "fusd"} {
		t.Error("invalid output for FUSD")
	}

	n = parseName("fusd")
	if n != [3]string{"FUSD", "FUSD", "fusd"} {
		t.Error("invalid output for fusd")
	}

	n = parseName("FlowToken")
	if n != [3]string{"FlowToken", "FLOW_TOKEN", "flowToken"} {
		t.Error("invalid output for FlowToken")
	}

	n = parseName("flow-token")
	if n != [3]string{"FlowToken", "FLOW_TOKEN", "flowToken"} {
		t.Error("invalid output for flow-token")
	}
}

func TestParseGenericFungibleTransfer(t *testing.T) {
	t.Run("FlowToken", func(t *testing.T) {
		g := ParseGenericFungibleTransfer(
			flow.Emulator,
			"FlowToken",
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
			"FlowToken",
			"some_other_tokenaddress", "some_other_baseaddress",
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
			"FUSD",
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
			"FUSD",
			"some_other_tokenaddress", "some_other_baseaddress",
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
