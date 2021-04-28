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

// List all accounts
func (s *AccountStore) Accounts() (accounts []data.Account, err error) {
	result := s.db.Find(&accounts)
	err = result.Error
	return
}

// Insert new account
func (s *AccountStore) InsertAccount(account data.Account) error {
	result := s.db.Create(&account)
	return result.Error
}

// Get account details
func (s *AccountStore) Account(address string) (account data.Account, err error) {
	result := s.db.First(&account, "address = ?", address)
	err = result.Error
	return
}

// List all account keys
func (s *AccountStore) AccountKeys() (keys []data.AccountKey, err error) {
	result := s.db.Find(&keys)
	err = result.Error
	return
}

// Insert new account key
func (s *AccountStore) InsertAccountKey(key data.AccountKey) error {
	result := s.db.Create(&key)
	return result.Error
}

// Get account key details
func (s *AccountStore) AccountKey(address string) (key data.AccountKey, err error) {
	result := s.db.First(&key, "address = ?", address)
	err = result.Error
	return
}
