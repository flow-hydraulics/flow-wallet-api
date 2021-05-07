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

func NewAccountStore(l *log.Logger, db *gorm.DB) (*AccountStore, error) {
	err := db.AutoMigrate(&data.Account{}, &data.Key{})
	if err != nil {
		return &AccountStore{}, err
	}
	return &AccountStore{l, db}, nil
}

func (s *AccountStore) Accounts() (accounts []data.Account, err error) {
	err = s.db.Find(&accounts).Error
	return
}

func (s *AccountStore) InsertAccount(account data.Account) error {
	return s.db.Create(&account).Error
}

func (s *AccountStore) Account(address string) (account data.Account, err error) {
	err = s.db.First(&account, "address = ?", address).Error
	return
}

func (s *AccountStore) AccountKey(address string) (key data.Key, err error) {
	err = s.db.Where(&data.Key{AccountAddress: address}).Order("updated_at asc").First(&key).Error
	s.db.Save(&key) // Update the UpdatedAt field
	return
}
