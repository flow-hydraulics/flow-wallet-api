package aws

import (
	"bytes"
	"os"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
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

	if testCfg.EncryptionKeyType != encryption.EncryptionKeyTypeAWSKMS {
		t.Skip("skipping since EncryptionKeyType is not", encryption.EncryptionKeyTypeAWSKMS)
	}
	// encrypt the test plaintext message
	crypter := AWSKMSCrypter{keyARN: testCfg.EncryptionKey}
	plaintext := []byte("this is a test message in plaintext")
	encrypted, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// decrypt the encrypted plaintext
	decrypted, err := crypter.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatal("decrypted does not match original plaintext message")
	}

}
