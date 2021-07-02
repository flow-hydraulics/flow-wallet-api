package accounts

import (
	"github.com/eqlabs/flow-wallet-api/datastore"
	"github.com/eqlabs/flow-wallet-api/keys"
	"github.com/eqlabs/flow-wallet-api/templates"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Account{}, &AccountToken{}, &keys.Storable{})
	return &GormStore{db}
}

func (s *GormStore) Accounts(o datastore.ListOptions) (aa []Account, err error) {
	err = s.db.
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&aa).Error
	return
}

func (s *GormStore) Account(address string) (a Account, err error) {
	err = s.db.First(&a, "address = ?", address).Error
	return
}

func (s *GormStore) InsertAccount(a *Account) error {
	return s.db.Create(a).Error
}

func (s *GormStore) AccountTokens(address string, tType *templates.TokenType) (att []AccountToken, err error) {
	q := s.db
	if tType != nil {
		// Filter by type
		q = q.Where(&AccountToken{AccountAddress: address, TokenType: *tType})
	} else {
		// Find all
		q = q.Where(&AccountToken{AccountAddress: address})
	}
	err = q.Order("token_name asc").Find(&att).Error
	return
}

func (s *GormStore) InsertAccountToken(at *AccountToken) error {
	return s.db.Create(at).Error
}
