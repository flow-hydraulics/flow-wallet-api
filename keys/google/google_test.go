package google

import (
	"context"
	"os"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
)

// Needs to be run manually with proper env configuration
// It's skipped during standard test execution
func TestGenerate(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	os.Setenv("FLOW_WALLET_ADMIN_ADDRESS", "0xf8d6e0586b0a20c7")
	os.Setenv("FLOW_WALLET_ADMIN_PRIVATE_KEY", "91a22fbd87392b019fbe332c32695c14cf2ba5b6521476a8540228bdf1987068")
	os.Setenv("FLOW_WALLET_ACCESS_API_HOST", "localhost:3569")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY", "faae4ed1c30f4e4555ee3a71f1044a8e")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY_TYPE", "local")

	testCfg, err := configs.Parse()
	if err != nil {
		t.Fatal(err)
	}

	// Safety measures
	testCfg.DatabaseDSN = "google_tests.db"
	testCfg.DatabaseType = "sqlite"
	testCfg.ChainID = flow.Emulator

	if testCfg.DefaultKeyType != keys.AccountKeyTypeGoogleKMS {
		t.Skip("skipping since DefaultKeyType is not", keys.AccountKeyTypeGoogleKMS)
	}

	t.Run("key is generated", func(t *testing.T) {
		flowAccountKey, privateKey, err := Generate(testCfg, context.Background(), 0, 1000)

		if err != nil {
			t.Fatal(err)
		}

		if flowAccountKey == nil {
			t.Fatal("Flow account key was not generated")
		}

		if privateKey == nil {
			t.Fatal("private key was not generated")
		}
	})
}
