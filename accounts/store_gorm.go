package accounts

import (
	"github.com/eqlabs/flow-wallet-service/keys"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Account{}, &keys.StorableKey{})
	return &GormStore{db}
}

func (s *GormStore) Accounts() (accounts []Account, err error) {
	err = s.db.Find(&accounts).Error
	return
}

func (s *GormStore) Account(address string) (account Account, err error) {
	err = s.db.First(&account, "address = ?", address).Error
	return
}

func (s *GormStore) InsertAccount(account Account) error {
	return s.db.Create(&account).Error
}
