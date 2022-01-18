package google

import (
	"bytes"
	"os"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
	"github.com/onflow/flow-go-sdk"
)

// Needs to be run manually with proper env configuration
// It's skipped during standard test execution
func TestCrypter(t *testing.T) {
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

	// Skip if not explicitly testing Google KMS keys
	if testCfg.EncryptionKeyType != encryption.EncryptionKeyTypeGoogleKMS {
		t.Skip("skipping since EncryptionKeyType is not", encryption.EncryptionKeyTypeGoogleKMS)
	}
	// Safety measures
	testCfg.DatabaseDSN = "google_tests.db"
	testCfg.DatabaseType = "sqlite"
	testCfg.ChainID = flow.Emulator

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
