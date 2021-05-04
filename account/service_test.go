package account

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/data/gorm"
	"github.com/eqlabs/flow-wallet-service/keys/simple"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type testConfig struct {
	FlowGateway string `env:"FLOW_GATEWAY,required"`
}

const testDbDSN = "test.db"
const testDbType = "sqlite"

func testSetup() (result *Service, err error) {
	godotenv.Load("../.env.test")

	os.Setenv("DB_DSN", testDbDSN)
	os.Setenv("DB_TYPE", testDbType)

	var cfg testConfig
	if err = env.Parse(&cfg); err != nil {
		return
	}

	l := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// Flow client
	fc, err := client.New(cfg.FlowGateway, grpc.WithInsecure())
	if err != nil {
		return
	}

	// Database
	db, err := gorm.NewStore(l)
	if err != nil {
		return
	}

	// Key manager
	km, err := simple.NewKeyManager(l, db, fc)
	if err != nil {
		return
	}

	result = NewService(l, db, km, fc)

	return
}

func testTearDown() {
	os.Remove(testDbDSN)
}

func TestAccountService(t *testing.T) {
	service, err := testSetup()
	if err != nil {
		t.Errorf("Error while running setup: %s", err)
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
		t.Errorf("Did not expect an error, got: %s", err)
		return
	}

	t.Run("new account has proper attributes", func(*testing.T) {
		if err := service.ValidateAddress(account.Address); err != nil {
			t.Errorf("Account has an invalid address: '%s'", account.Address)
		}

		if len(account.Keys) != 0 {
			t.Errorf("Account should not expose keys")
		}
	})

	t.Run("account can make a transaction", func(t *testing.T) {})

	testTearDown()
}
