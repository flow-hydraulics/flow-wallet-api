// Package encryption provides encryption and decryption.
package encryption

type Crypter interface {
	Encrypt(message []byte) (encrypted []byte, err error)
	Decrypt(encrypted []byte) (message []byte, err error)
}

const EncryptionKeyTypeGoogleKMS = "google_kms"
const EncryptionKeyTypeAWSKMS = "aws_kms"
const EncryptionKeyTypeLocal = "local"
