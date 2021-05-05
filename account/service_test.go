package account

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/tokens"
	"github.com/onflow/flow-go-sdk"
)

func TestAccountService(t *testing.T) {
	l := log.New(ioutil.Discard, "", log.LstdFlags|log.Lshortfile)
	service, err := SetupTestService(l)
	if err != nil {
		t.Fatalf("Error while running setup: %s", err)
	}

	t.Run("ValidateAddress", func(t *testing.T) {
		if err := service.ValidateAddress("not-a-valid-address"); err == nil {
			t.Errorf("Expected an error")
		}

		if err := service.ValidateAddress(""); err == nil {
			t.Errorf("Expected an error")
		}
	})

	account, err := service.Create(context.Background())
	if err != nil {
		t.Fatalf("Did not expect an error, got: %s", err)
	}

	t.Run("new account has proper attributes", func(*testing.T) {
		if err := service.ValidateAddress(account.Address); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Errorf("Account should not expose keys")
		}
	})

	t.Run("account can make a transaction", func(t *testing.T) {
		// Fund the account from service account
		tokens.TransferFlow(
			service.km,
			service.fc,
			flow.HexToAddress(account.Address),
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			"1.0",
		)

		txId, err := tokens.TransferFlow(
			service.km,
			service.fc,
			flow.HexToAddress(os.Getenv("ADMIN_ACC_ADDRESS")),
			flow.HexToAddress(account.Address),
			"1.0",
		)

		if err != nil {
			t.Fatalf("Did not expect an error, got: %s", err)
		}

		if txId == flow.EmptyID {
			t.Fatalf("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), service.fc, txId)
		if err != nil {
			t.Errorf("Did not expect an error, got: %s", err)
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
			t.Fatalf("Did not expect an error, got: %s", err)
		}

		if txId == flow.EmptyID {
			t.Fatalf("Expected txId not to be empty")
		}

		_, err = flow_helpers.WaitForSeal(context.Background(), service.fc, txId)
		if err == nil {
			t.Errorf("Expected an error")
		}
	})

	TeardDownTestService()
}
