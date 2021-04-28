package memory

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
)

type AccountStore struct {
	StoredAccounts    map[string]data.Account
	StoredAccountKeys map[string]data.AccountKey
}

func newAccountStore() *AccountStore {
	return &AccountStore{
		StoredAccounts:    make(map[string]data.Account),
		StoredAccountKeys: make(map[string]data.AccountKey),
	}
}

func (s *AccountStore) Accounts() ([]data.Account, error) {
	values := []data.Account{}
	for _, value := range s.StoredAccounts {
		values = append(values, value)
	}
	return values, nil
}

func (s *AccountStore) Account(address string) (data.Account, error) {
	return s.StoredAccounts[address], nil
}

func (s *AccountStore) InsertAccount(a data.Account) error {
	s.StoredAccounts[a.Address] = a
	return nil
}

func (s *AccountStore) DeleteAccount(address string) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKeys() ([]data.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKey(address string) (data.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccountKey(k data.AccountKey) error {
	s.StoredAccountKeys[k.AccountAddress] = k
	return nil
}

func (s *AccountStore) DeleteAccountKey(address string) error {
	panic("not implemented") // TODO: Implement
}
