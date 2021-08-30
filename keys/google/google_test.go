package google

import (
	"context"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
)

func TestGenerate(t *testing.T) {
	opts := &configs.Options{EnvFilePath: "../../.env.test"}
	testCfg, err := configs.ParseConfig(opts)

	if err != nil {
		t.Fatal(err)
	}

	if testCfg.DefaultKeyType != keys.AccountKeyTypeGoogleKMS {
		t.Skip("skipping since default key type is not", keys.AccountKeyTypeGoogleKMS)
	}

	t.Run("key is generated", func(t *testing.T) {
		flowAccountKey, privateKey, err := Generate(testCfg, context.TODO(), 0, 1000)

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
