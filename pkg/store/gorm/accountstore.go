package gorm

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"gorm.io/gorm"
)

type AccountStore struct {
	db *gorm.DB
}

func newAccountStore(db *gorm.DB) *AccountStore {
	db.AutoMigrate(&store.Account{}, &store.AccountKey{})
	return &AccountStore{db}
}

func (s *AccountStore) Accounts() ([]store.Account, error) {
	var accounts []store.Account
	result := s.db.Find(&accounts)
	return accounts, result.Error
}

func (s *AccountStore) Account(address string) (store.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccount(a store.Account) error {
	result := s.db.Create(&a)
	return result.Error
}

func (s *AccountStore) DeleteAccount(address string) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKeys() ([]store.AccountKey, error) {
	var keys []store.AccountKey
	result := s.db.Find(&keys)
	return keys, result.Error
}

func (s *AccountStore) AccountKey(address string) (store.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccountKey(k store.AccountKey) error {
	result := s.db.Create(&k)
	return result.Error
}

func (s *AccountStore) DeleteAccountKey(address string) error {
	panic("not implemented") // TODO: Implement
}
