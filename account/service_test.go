package account

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/tokens"
	"github.com/onflow/flow-go-sdk"
)

func TestAccountService(t *testing.T) {
	l := log.New(NewTestLogger(t), "", log.LstdFlags|log.Lshortfile)

	service, err := TestServiceSetup(l)
	if err != nil {
		t.Fatalf("Error while running setup: %s", err)
	}

	t.Run("ValidateAddress", func(t *testing.T) {
		if err := service.ValidateAddress("not-a-valid-address"); err == nil {
			t.Error("Expected an error")
		}

		if err := service.ValidateAddress(""); err == nil {
			t.Error("Expected an error")
		}
	})

	account, err := service.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("new account has proper attributes", func(*testing.T) {
		if err := service.ValidateAddress(account.Address); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Error("Account should not expose keys")
		}
	})

	t.Run("account can make a transaction", func(t *testing.T) {
		// Fund the account from service account
		txId, err := tokens.TransferFlow(
			service.km,
			service.fc,
			flow.HexToAddress(account.Address),
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			"1.0",
		)
		if err != nil {
			t.Fatal(err)
		}
		_, err = flow_helpers.WaitForSeal(context.Background(), service.fc, txId)
		if err != nil {
			t.Fatal(err)
		}

		txId, err = tokens.TransferFlow(
			service.km,
			service.fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(account.Address),
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		if txId == flow.EmptyID {
			t.Fatalf("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), service.fc, txId)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("account can not make a transaction without funds", func(t *testing.T) {
		txId, err := tokens.TransferFlow(
			service.km,
			service.fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(account.Address),
			"1.0",
		)

		if err != nil {
			t.Fatal(err)
		}

		if txId == flow.EmptyID {
			t.Fatal("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), service.fc, txId)
		if err == nil {
			t.Fatal("Expected an error")
		}
	})

	TestServiceTearDown()
}
