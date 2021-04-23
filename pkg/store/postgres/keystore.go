package postgres

import "github.com/eqlabs/flow-nft-wallet-service/pkg/store"

type KeyStore struct {
	store.KeyStore
}

func NewKeyStore() (*KeyStore, error) {
	return &KeyStore{}, nil
}
