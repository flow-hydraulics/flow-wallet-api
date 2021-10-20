package tests

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

var max_tx_wait = 10 * time.Second

func Test_NonCustodialAccountDepositTracking(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	fc := test.NewFlowClient(t, cfg)
	svcs := test.GetServices(t, cfg)
	accountSvc := svcs.GetAccounts()
	km := svcs.GetKeyManager()

	adminAuthorizer, err := km.AdminAuthorizer(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	nonCustodialAccount := test.NewFlowAccount(t, fc, adminAuthorizer.Address, adminAuthorizer.Key, adminAuthorizer.Signer)

	_, custodialAccount, err := accountSvc.Create(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("non-custodial account: %q", nonCustodialAccount.Address.Hex())
	t.Logf("    custodial account: %q", custodialAccount.Address)

	_, err = accountSvc.AddNonCustodialAccount(context.Background(), nonCustodialAccount.Address.Hex())
	if err != nil {
		t.Fatal(err)
	}

	deposits, err := svcs.GetTokens().ListDeposits("0x"+nonCustodialAccount.Address.Hex(), "FlowToken")
	if err != nil {
		t.Fatal(err)
	}

	if len(deposits) != 0 {
		// There shouldn't be any before token transfers.
		t.Fatalf("got %d deposits for non-custodial account, expected 0", len(deposits))
	}

	verifyBalance(t, fc, flow.HexToAddress(custodialAccount.Address), 100000)
	verifyBalance(t, fc, nonCustodialAccount.Address, 100000)

	transferTokens(t, context.Background(), fc, km, "100.0", cfg.AdminAddress, custodialAccount.Address)
	transferTokens(t, context.Background(), fc, km, "10.0", custodialAccount.Address, nonCustodialAccount.Address.Hex())

	verifyBalance(t, fc, flow.HexToAddress(custodialAccount.Address), 9000100000)
	verifyBalance(t, fc, nonCustodialAccount.Address, 1000100000)

	// The chain event tracking is a background goroutine which runs every now
	// and then. Test must wait for a little while in order to allow event
	// tracking to see & process the token deposit.
	time.Sleep(time.Second)

	deposits, err = svcs.GetTokens().ListDeposits("0x"+nonCustodialAccount.Address.Hex(), "FlowToken")
	if err != nil {
		t.Fatal(err)
	}

	if len(deposits) == 0 {
		t.Fatal("got 0 deposits for non-custodial account, expected more than 0")
	}
}

func transferTokens(t *testing.T, ctx context.Context, fc *client.Client, km keys.Manager, amount, proposerAddr, receiverAddr string) {
	t.Helper()

	amountArg, err := cadence.NewUFix64(amount)
	if err != nil {
		t.Fatal(err)
	}

	lastBlock, err := fc.GetLatestBlockHeader(ctx, true)
	proposer, err := fc.GetAccountAtLatestBlock(ctx, flow.HexToAddress(proposerAddr))
	seqNum := proposer.Keys[0].SequenceNumber

	payer, err := km.AdminAuthorizer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	authorizer, err := km.UserAuthorizer(ctx, proposer.Address)
	if err != nil {
		t.Fatal(err)
	}
	tx := flow.NewTransaction()
	tx.SetScript(test.ReadFile(t, "fixtures/transfer_tokens.cdc"))
	tx.AddArgument(amountArg)
	tx.AddArgument(cadence.NewAddress(flow.HexToAddress(receiverAddr)))
	tx.SetGasLimit(9999)
	tx.SetReferenceBlockID(lastBlock.ID)
	tx.SetProposalKey(proposer.Address, 0, seqNum)
	tx.SetPayer(payer.Address)
	tx.AddAuthorizer(authorizer.Address)

	if authorizer.Address != payer.Address {
		tx.SignPayload(authorizer.Address, authorizer.Key.Index, authorizer.Signer)
	}
	tx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer)
	err = fc.SendTransaction(context.Background(), *tx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = flow_helpers.WaitForSeal(ctx, fc, tx.ID(), max_tx_wait)
	if err != nil {
		t.Fatal(err)
	}
}

func verifyBalance(t *testing.T, fc *client.Client, address flow.Address, expected uint64) {
	t.Helper()

	a, err := fc.GetAccount(context.Background(), address)
	if err != nil {
		t.Fatal(err)
	}

	if a.Balance != expected {
		t.Fatalf("expected account %q balance to be %d, got %d", address.Hex(), expected, a.Balance)
	}
}
