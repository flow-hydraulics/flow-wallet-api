package postgres

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
)

type DataStore struct {
	store.AccountStore
}

type AccountStore struct{}

func NewDataStore() (*DataStore, error) {
	return &DataStore{
		AccountStore: newAccountStore(),
	}, nil
}

func newAccountStore() *AccountStore {
	return &AccountStore{}
}

func (s *AccountStore) Account(address flow.Address) (store.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccount(a store.Account) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) DeleteAccount(address flow.Address) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKey(address flow.Address) (store.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccountKey(k store.AccountKey) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) DeleteAccountKey(address flow.Address) error {
	panic("not implemented") // TODO: Implement
}
