package postgres

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/google/uuid"
)

type DataStore struct {
	store.DataStore
}

func NewDataStore() (*DataStore, error) {
	return &DataStore{}, nil
}

func (s *DataStore) Account(id uuid.UUID) (store.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DataStore) Accounts() ([]store.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DataStore) Transaction(id uuid.UUID) (store.Transaction, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DataStore) Transactions() ([]store.Transaction, error) {
	panic("not implemented") // TODO: Implement
}
