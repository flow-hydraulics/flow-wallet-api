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
	var err error

	opts := &configs.Options{EnvFilePath: "../../.env.test"}
	testCfg, err = configs.ParseConfig(opts)
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
