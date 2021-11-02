package aws

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
	if err != nil {
		t.Fatal(err)
	}

	// Safety measures
	testCfg.DatabaseDSN = "aws_tests.db"
	testCfg.DatabaseType = "sqlite"
	testCfg.ChainID = flow.Emulator

	// Skip if not explicitly testing AWS KMS keys
	if testCfg.EncryptionKeyType != encryption.EncryptionKeyTypeAWSKMS {
		t.Skip("skipping since EncryptionKeyType is not", encryption.EncryptionKeyTypeAWSKMS)
	}

	// encrypt
	crypter := AWSKMSCrypter{keyARN: testCfg.EncryptionKey}
	message := []byte("this is a test message")
	encrypted, err := crypter.Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}

	// decrypt
	decrypted, err := crypter.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, message) {
		t.Fatal("decrypted does not match original message")
	}

}
