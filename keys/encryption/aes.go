package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AESCrypter struct {
	key []byte
}

func NewAESCrypter(key []byte) *AESCrypter {
	return &AESCrypter{key}
}

func (s *AESCrypter) Encrypt(message []byte) ([]byte, error) {
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

	return gcm.Seal(nonce, nonce, message, nil), nil
}

func (s *AESCrypter) Decrypt(encrypted []byte) ([]byte, error) {
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
	if len(encrypted) < nonceSize {
		return []byte(""), fmt.Errorf("message too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return []byte(""), err
	}

	return plaintext, nil
}
