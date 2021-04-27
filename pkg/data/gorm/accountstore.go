package gorm

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"gorm.io/gorm"
)

type AccountStore struct {
	db *gorm.DB
}

func newAccountStore(db *gorm.DB) *AccountStore {
	db.AutoMigrate(&data.Account{}, &data.AccountKey{})
	return &AccountStore{db}
}

func (s *AccountStore) Accounts() ([]data.Account, error) {
	var accounts []data.Account
	result := s.db.Find(&accounts)
	return accounts, result.Error
}

func (s *AccountStore) Account(address string) (data.Account, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccount(a data.Account) error {
	result := s.db.Create(&a)
	return result.Error
}

func (s *AccountStore) DeleteAccount(address string) error {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) AccountKeys() ([]data.AccountKey, error) {
	var keys []data.AccountKey
	result := s.db.Find(&keys)
	return keys, result.Error
}

func (s *AccountStore) AccountKey(address string) (data.AccountKey, error) {
	panic("not implemented") // TODO: Implement
}

func (s *AccountStore) InsertAccountKey(k data.AccountKey) error {
	result := s.db.Create(&k)
	return result.Error
}

func (s *AccountStore) DeleteAccountKey(address string) error {
	panic("not implemented") // TODO: Implement
}
