package gorm

import (
	"log"

	"github.com/eqlabs/flow-wallet-service/data"
	"gorm.io/gorm"
)

type AccountStore struct {
	l  *log.Logger
	db *gorm.DB
}

func newAccountStore(l *log.Logger, db *gorm.DB) *AccountStore {
	db.AutoMigrate(&data.Account{}, &data.Key{})
	return &AccountStore{l, db}
}

// List all accounts
func (s *AccountStore) Accounts() (accounts []data.Account, err error) {
	err = s.db.Find(&accounts).Error
	return
}

// Insert new account
func (s *AccountStore) InsertAccount(account data.Account) error {
	return s.db.Create(&account).Error
}

// Get account details
func (s *AccountStore) Account(address string) (account data.Account, err error) {
	err = s.db.First(&account, "address = ?", address).Error
	return
}

// Get account key with index
func (s *AccountStore) AccountKey(address string, index int) (key data.Key, err error) {
	err = s.db.Where("account_address = ? AND index = ?", address, index).First(&key).Error
	return
}
