package aws

import (
	"bytes"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
)

func TestCrypter(t *testing.T) {
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
