package simple

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type SymmetricCrypter struct {
	key []byte
}

func NewCrypter(key []byte) *SymmetricCrypter {
	return &SymmetricCrypter{key}
}

func (s *SymmetricCrypter) Encrypt(original []byte) ([]byte, error) {
	// TODO: research if these could be stored in the struct
	c, err := aes.NewCipher(s.key)
	if err != nil {
		return []byte(""), err
	}

	// TODO: research if these could be stored in the struct
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte(""), err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte(""), err
	}

	return gcm.Seal(nonce, nonce, original, nil), nil
}

func (s *SymmetricCrypter) Decrypt(original []byte) ([]byte, error) {
	// TODO: research if these could be stored in the struct
	c, err := aes.NewCipher(s.key)
	if err != nil {
		return []byte(""), err
	}

	// TODO: research if these could be stored in the struct
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte(""), err
	}

	nonceSize := gcm.NonceSize()
	if len(original) < nonceSize {
		return []byte(""), fmt.Errorf("message too short")
	}

	nonce, ciphertext := original[:nonceSize], original[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return []byte(""), err
	}

	return plaintext, nil
}
