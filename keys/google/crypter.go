package google

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"io"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GoogleKMSCrypter struct {
	keyResourceName string
}

// NewGoogleKMSCrypter creates a new GoogleKMSCrypter
// with the specified key (KMS key resource name)
func NewGoogleKMSCrypter(key []byte) *GoogleKMSCrypter {
	return &GoogleKMSCrypter{keyResourceName: string(key)}
}

// Encrypt encrypts the given data with the symmetric encryption key
// specified in the crypter
func (c *GoogleKMSCrypter) Encrypt(message []byte) (encrypted []byte, err error) {
	ctx := context.Background()

	res := new(bytes.Buffer)
	err = encryptSymmetric(ctx, res, c.keyResourceName, message)
	if err != nil {
		return encrypted, err
	}

	encrypted = res.Bytes()

	return encrypted, err
}

// Decrypt decryptes the given encrypted data with the symmetric encryption key
// specified in the crypter
func (c *GoogleKMSCrypter) Decrypt(encrypted []byte) (message []byte, err error) {
	ctx := context.Background()

	res := new(bytes.Buffer)
	err = decryptSymmetric(ctx, res, c.keyResourceName, encrypted)
	if err != nil {
		return message, err
	}

	message = res.Bytes()

	return message, err
}

// encryptSymmetric encrypts the input plaintext with the specified symmetric KMS key
func encryptSymmetric(ctx context.Context, w io.Writer, name string, plaintext []byte) error {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kms client: %v", err)
	}
	defer client.Close()

	plaintextCRC32C := crc32c(plaintext)

	result, err := client.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:            name,
		Plaintext:       plaintext,
		PlaintextCrc32C: wrapperspb.Int64(int64(plaintextCRC32C)),
	})
	if err != nil {
		return fmt.Errorf("failed to encrypt: %v", err)
	}

	// Verify result integrity
	if !result.VerifiedPlaintextCrc32C {
		return fmt.Errorf("Encrypt: request corrupted in-transit")
	}
	if int64(crc32c(result.Ciphertext)) != result.CiphertextCrc32C.Value {
		return fmt.Errorf("Encrypt: response corrupted in-transit")
	}

	_, err = w.Write(result.Ciphertext)
	if err != nil {
		return fmt.Errorf("failed to write encrypted value: %v", err)
	}

	return nil
}

// decryptSymmetric will decrypt the input ciphertext bytes using the specified symmetric KMS key
func decryptSymmetric(ctx context.Context, w io.Writer, name string, ciphertext []byte) error {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kms client: %v", err)
	}
	defer client.Close()

	ciphertextCRC32C := crc32c(ciphertext)

	result, err := client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:             name,
		Ciphertext:       ciphertext,
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	})
	if err != nil {
		return fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}

	// Verify result integrity
	if int64(crc32c(result.Plaintext)) != result.PlaintextCrc32C.Value {
		return fmt.Errorf("Decrypt: response corrupted in-transit")
	}

	_, err = w.Write(result.Plaintext)
	if err != nil {
		return fmt.Errorf("failed to write decrypted value: %v", err)
	}

	return nil
}

// crc32c calculates a crc32 checksum on the given data
func crc32c(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}
