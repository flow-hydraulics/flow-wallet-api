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

func NewGoogleKMSCrypter(key []byte) *GoogleKMSCrypter {
	return &GoogleKMSCrypter{keyResourceName: string(key)}
}

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

// encryptSymmetric encrypts the input plaintext with the specified symmetric
// Cloud KMS key.
func encryptSymmetric(ctx context.Context, w io.Writer, name string, plaintext []byte) error {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kms client: %v", err)
	}
	defer client.Close()

	// Optional but recommended: Compute plaintext's CRC32C.
	plaintextCRC32C := crc32c(plaintext)

	// Build the request.
	req := &kmspb.EncryptRequest{
		Name:            name,
		Plaintext:       plaintext,
		PlaintextCrc32C: wrapperspb.Int64(int64(plaintextCRC32C)),
	}

	// Call the API.
	result, err := client.Encrypt(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %v", err)
	}

	// Optional, but recommended: perform integrity verification on result.
	// For more details on ensuring E2E in-transit integrity to and from Cloud KMS visit:
	// https://cloud.google.com/kms/docs/data-integrity-guidelines
	if !result.VerifiedPlaintextCrc32C {
		return fmt.Errorf("Encrypt: request corrupted in-transit")
	}
	if int64(crc32c(result.Ciphertext)) != result.CiphertextCrc32C.Value {
		return fmt.Errorf("Encrypt: response corrupted in-transit")
	}

	w.Write(result.Ciphertext)
	return nil
}

// decryptSymmetric will decrypt the input ciphertext bytes using the specified symmetric key.
func decryptSymmetric(ctx context.Context, w io.Writer, name string, ciphertext []byte) error {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create kms client: %v", err)
	}
	defer client.Close()

	// Optional, but recommended: Compute ciphertext's CRC32C.
	ciphertextCRC32C := crc32c(ciphertext)

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:             name,
		Ciphertext:       ciphertext,
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	}

	// Call the API.
	result, err := client.Decrypt(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to decrypt ciphertext: %v", err)
	}

	// Optional, but recommended: perform integrity verification on result.
	// For more details on ensuring E2E in-transit integrity to and from Cloud KMS visit:
	// https://cloud.google.com/kms/docs/data-integrity-guidelines
	if int64(crc32c(result.Plaintext)) != result.PlaintextCrc32C.Value {
		return fmt.Errorf("Decrypt: response corrupted in-transit")
	}

	w.Write(result.Plaintext)
	return nil
}

func crc32c(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}
