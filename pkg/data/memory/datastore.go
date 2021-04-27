package memory

import "github.com/eqlabs/flow-nft-wallet-service/pkg/data"

type DataStore struct {
	data.AccountStore
}

func NewDataStore() (*DataStore, error) {
	return &DataStore{
		AccountStore: newAccountStore(),
	}, nil
}
