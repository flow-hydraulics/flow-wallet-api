package aws

import (
	"context"
	"os"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

var testCfg *configs.Config

func TestMain(m *testing.M) {
	os.Clearenv()
	defer os.Clearenv()
	os.Setenv("FLOW_WALLET_ADMIN_ADDRESS", "0xf8d6e0586b0a20c7")
	os.Setenv("FLOW_WALLET_ADMIN_PRIVATE_KEY", "91a22fbd87392b019fbe332c32695c14cf2ba5b6521476a8540228bdf1987068")
	os.Setenv("FLOW_WALLET_ACCESS_API_HOST", "localhost:3569")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY", "faae4ed1c30f4e4555ee3a71f1044a8e")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY_TYPE", "local")

	var err error
	testCfg, err = configs.Parse()
	if err != nil {
		log.
			WithFields(log.Fields{"error": err}).
			Warn("Could not parse config")
		os.Exit(1)
	}

	// Safety measures
	testCfg.DatabaseDSN = "aws_tests.db"
	testCfg.DatabaseType = "sqlite"
	testCfg.ChainID = flow.Emulator

	exitVal := m.Run()

	os.Exit(exitVal)
}

// Needs to be run manually with proper env configuration
// It's skipped during standard test execution
func TestGenerate(t *testing.T) {
	if testCfg.DefaultKeyType != keys.AccountKeyTypeAWSKMS {
		t.Skip("skipping since DefaultKeyType is not", keys.AccountKeyTypeAWSKMS)
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
