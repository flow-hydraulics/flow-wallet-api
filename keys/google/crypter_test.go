package google

import (
	"bytes"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
)

// Needs to be run manually with proper env configuration
// It's skipped during standard test execution
func TestCrypter(t *testing.T) {
	cfg := configs.ParseTestConfig(t)

	// Skip if not explicitly testing Google KMS keys
	if cfg.EncryptionKeyType != encryption.EncryptionKeyTypeGoogleKMS {
		t.Skip("skipping since EncryptionKeyType is not", encryption.EncryptionKeyTypeGoogleKMS)
	}

	// Encrypt example test message with configured encryption key (key resource name)
	crypter := GoogleKMSCrypter{keyResourceName: cfg.EncryptionKey}
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
