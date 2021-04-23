package memory

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/google/uuid"
)

type DataStore struct {
	store.AccountStore
}

type AccountStore struct {
	StoredAccounts map[uuid.UUID]store.Account
}

func NewDataStore() (*DataStore, error) {
	return &DataStore{
		AccountStore: NewAccountStore(),
	}, nil
}

func NewAccountStore() *AccountStore {
	return &AccountStore{StoredAccounts: make(map[uuid.UUID]store.Account)}
}

func (s *AccountStore) Account(id uuid.UUID) (store.Account, error) {
	return s.StoredAccounts[id], nil
}

func (s *AccountStore) Accounts() ([]store.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) CreateAccount(a *store.Account) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) DeleteAccount(id uuid.UUID) error {
	panic("not implemented") // TODO: Implement
}
