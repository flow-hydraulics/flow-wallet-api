package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/tests/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/onflow/flow-go-sdk"
)

func Test_TransactionSignByAdmin(t *testing.T) {
	cfg := test.LoadConfig(t)
	txSvc := test.GetServices(t, cfg).GetTransactions()

	tx, err := txSvc.Sign(context.Background(), cfg.AdminAddress, "", nil)
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
	cfg := test.LoadConfig(t)
	svcs := test.GetServices(t, cfg)

	_, acc, err := svcs.GetAccounts().Create(ctx, true)
	if err != nil {
		t.Fatalf("expected err == nil, got %#v", err)
	}

	tx, err := svcs.GetTransactions().Sign(ctx, acc.Address, "", nil)
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

func Test_TransactionProposalKeySequenceNumber(t *testing.T) {
	t.Run("update without re-signing", func(t *testing.T) {
		cfg := test.LoadConfig(t)
		app := test.GetServices(t, cfg)
		txSvc := app.GetTransactions()
		system := app.GetSystem()

		// Pause, so the transaction won't get automatically sent
		if err := system.Pause(); err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()

		_, tx, err := txSvc.Create(ctx, false, cfg.AdminAddress, "transaction() { prepare(signer: AuthAccount){} execute {}}", nil, transactions.General)
		if err != nil {
			t.Fatal(err)
		}

		flowTx, err := flow.DecodeTransaction(tx.FlowTransaction)
		if err != nil {
			t.Fatal(err)
		}

		// Update the sequence number
		flowTx.
			SetProposalKey(flowTx.ProposalKey.Address, flowTx.ProposalKey.KeyIndex, flowTx.ProposalKey.SequenceNumber+1)

		// Should return a "signature is not valid" error
		_, err = flow_helpers.SendAndWait(ctx, app.GetFlowClient(), *flowTx, 0)
		if err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("update sequence number during job run", func(t *testing.T) {
		t.Skip("not supported currently")

		cfg := test.LoadConfig(t)
		cfg.AdminProposalKeyCount = 1
		cfg.WorkerCount = 1
		app := test.GetServices(t, cfg)
		txSvc := app.GetTransactions()
		system := app.GetSystem()

		// Pause, so the transactions won't get immediately sent
		if err := system.Pause(); err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()

		job1, tx1, err := txSvc.Create(ctx, false, cfg.AdminAddress, "transaction() { prepare(signer: AuthAccount){} execute {}}", nil, transactions.General)
		if err != nil {
			t.Fatal(err)
		}

		job2, tx2, err := txSvc.Create(ctx, false, cfg.AdminAddress, "transaction() { prepare(signer: AuthAccount){} execute {}}", nil, transactions.General)
		if err != nil {
			t.Fatal(err)
		}

		flowTx1, err := flow.DecodeTransaction(tx1.FlowTransaction)
		if err != nil {
			t.Fatal(err)
		}

		flowTx2, err := flow.DecodeTransaction(tx2.FlowTransaction)
		if err != nil {
			t.Fatal(err)
		}

		if flowTx1.ProposalKey.KeyIndex != flowTx2.ProposalKey.KeyIndex {
			t.Fatal("expected key indexes to match")
		}

		if flowTx1.ProposalKey.SequenceNumber != flowTx2.ProposalKey.SequenceNumber {
			t.Fatal("expected sequence numbers to match")
		}

		// Resume when both transactions are in the queue
		if err := system.Resume(); err != nil {
			t.Fatal(err)
		}

		if _, err := test.WaitForJob(app.GetJobs(), job1.ID.String()); err != nil {
			t.Error(err)
		}

		if _, err := test.WaitForJob(app.GetJobs(), job2.ID.String()); err != nil {
			t.Error(err)
		}
	})

}
