package kms

import (
	"github.com/onflow/flow-go-sdk/client"
)

type KeyStore struct {
	fc *client.Client
}

func NewKeyStore(fc *client.Client) (*KeyStore, error) {
	return &KeyStore{fc}, nil
}
