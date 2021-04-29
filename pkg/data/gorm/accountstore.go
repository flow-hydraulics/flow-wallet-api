package gorm

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"gorm.io/gorm"
)

type AccountStore struct {
	db *gorm.DB
}

func newAccountStore(db *gorm.DB) *AccountStore {
	db.AutoMigrate(&data.Account{}, &data.Key{})
	return &AccountStore{db}
}

// List all accounts
func (s *AccountStore) Accounts() (accounts []data.Account, err error) {
	err = s.db.Select("address").Find(&accounts).Error
	return
}

// Insert new account
func (s *AccountStore) InsertAccount(account data.Account) error {
	return s.db.Create(&account).Error
}

// Get account details
func (s *AccountStore) Account(address string) (account data.Account, err error) {
	err = s.db.Preload("Keys").First(&account, "address = ?", address).Error
	return
}

// Get account key with index
func (s *AccountStore) AccountKey(address string, index int) (key data.Key, err error) {
	err = s.db.Where("account_address = ? AND index = ?", address, index).First(&key).Error
	return
}
