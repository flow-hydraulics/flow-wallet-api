package google

import (
	"bytes"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
	"github.com/onflow/flow-go-sdk"
)

func TestCrypter(t *testing.T) {
	opts := &configs.Options{EnvFilePath: "../../.env.test"}
	testCfg, err := configs.ParseConfig(opts)

	// Skip if not explicitly testing Google KMS keys
	if testCfg.EncryptionKeyType != encryption.EncryptionKeyTypeGoogleKMS {
		t.Skip("skipping since EncryptionKeyType is not", encryption.EncryptionKeyTypeGoogleKMS)
	}

	// Safety measures
	testCfg.DatabaseDSN = "google_tests.db"
	testCfg.DatabaseType = "sqlite"
	testCfg.ChainID = flow.Emulator

	if err != nil {
		t.Fatal(err)
	}

	// Encrypt example test message with configured encryption key (key resource name)
	crypter := GoogleKMSCrypter{keyResourceName: testCfg.EncryptionKey}
	message := []byte("this is a test message")
	encrypted, err := crypter.Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}

	// Decrypt the newly encrypted test message
	decrypted, err := crypter.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, message) {
		t.Fatal("decrypted does not match original message")
	}

}
