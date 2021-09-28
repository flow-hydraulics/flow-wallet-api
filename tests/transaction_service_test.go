package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/onflow/flow-go-sdk"
)

func Test_TransactionSignByAdmin(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	txSvc := test.GetServices(t, cfg).GetTransactions()

	tx, err := txSvc.Sign(context.Background(), cfg.AdminAddress, templates.Raw{}, transactions.General)
	if err != nil {
		t.Fatalf("expected err == nil, got %#v", err)
	}

	// Verify that when payer == authorizer, there's no payload envelope signatures.
	if len(tx.PayloadSignatures) > 0 {
		t.Fatalf("expected len(tx.PayloadSignatures) == 0, got %d", len(tx.PayloadSignatures))
	}

	// Verify presence of payment envelope signature.
	if len(tx.EnvelopeSignatures) == 0 {
		t.Fatal("expected len(tx.EnvelopeSignatures) > 0, got 0")
	}

	if !addressExists(cfg.AdminAddress, tx.EnvelopeSignatures) {
		t.Fatalf("couldn't find signer's address from envelope signatures")
	}
}

func Test_TransactionSignByAnotherAccount(t *testing.T) {
	ctx := context.Background()
	cfg := test.LoadConfig(t, testConfigPath)
	svcs := test.GetServices(t, cfg)

	_, acc, err := svcs.GetAccounts().Create(ctx, true)
	if err != nil {
		t.Fatalf("expected err == nil, got %#v", err)
	}

	tx, err := svcs.GetTransactions().Sign(ctx, acc.Address, templates.Raw{}, transactions.General)
	if err != nil {
		t.Fatalf("expected err == nil, got %#v", err)
	}

	// Verify that when payer != authorizer, there's payload envelope signature[s] present.
	if len(tx.PayloadSignatures) == 0 {
		t.Fatal("expected len(tx.PayloadSignatures) > 0, got 0")
	}

	// Verify that payload signature is signed by correct address.
	if !addressExists(acc.Address, tx.PayloadSignatures) {
		t.Fatalf("couldn't find signer's address from payload signatures")
	}

	// Verify presence of payment envelope signature[s].
	if len(tx.EnvelopeSignatures) == 0 {
		t.Fatal("expected len(tx.EnvelopeSignatures) > 0, got 0")
	}
}

func addressExists(addr string, sigs []flow.TransactionSignature) bool {
	addr = strings.TrimPrefix(addr, "0x")
	for _, s := range sigs {
		if s.Address.Hex() == addr {
			return true
		}
	}

	return false
}
