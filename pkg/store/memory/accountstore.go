package memory

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
)

type AccountStore struct {
	StoredAccounts    map[string]store.Account
	StoredAccountKeys map[string]store.AccountKey
}

func newAccountStore() *AccountStore {
	return &AccountStore{
		StoredAccounts:    make(map[string]store.Account),
		StoredAccountKeys: make(map[string]store.AccountKey),
	}
}

func (s *AccountStore) Accounts() ([]store.Account, error) {
	values := []store.Account{}
	for _, value := range s.StoredAccounts {
		values = append(values, value)
	}
	return values, nil
}

func (s *AccountStore) Account(address string) (store.Account, error) {
	return s.StoredAccounts[address], nil
}

func (s *AccountStore) InsertAccount(a store.Account) error {
	s.StoredAccounts[a.Address] = a
	return nil
}

func (s *AccountStore) DeleteAccount(address string) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKeys() ([]store.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKey(address string) (store.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccountKey(k store.AccountKey) error {
	s.StoredAccountKeys[k.AccountAddress] = k
	return nil
}

func (s *AccountStore) DeleteAccountKey(address string) error {
	panic("not implemented") // TODO: Implement
}
