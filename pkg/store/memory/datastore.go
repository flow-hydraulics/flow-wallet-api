package memory

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
)

type DataStore struct {
	store.AccountStore
}

type AccountStore struct {
	StoredAccounts map[flow.Address]store.Account
}

func NewDataStore() (*DataStore, error) {
	return &DataStore{
		AccountStore: newAccountStore(),
	}, nil
}

func newAccountStore() *AccountStore {
	return &AccountStore{StoredAccounts: make(map[flow.Address]store.Account)}
}

func (s *AccountStore) Account(address flow.Address) (store.Account, error) {
	return s.StoredAccounts[address], nil
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

func (s *AccountStore) InsertAccountKey(a store.AccountKey) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) DeleteAccountKey(address flow.Address) error {
	panic("not implemented") // TODO: Implement
}
