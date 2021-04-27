package memory

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
)

type DataStore struct {
	store.AccountStore
}

func NewDataStore() (*DataStore, error) {
	return &DataStore{
		AccountStore: newAccountStore(),
	}, nil
}
