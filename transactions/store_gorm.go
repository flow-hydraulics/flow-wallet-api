package transactions

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Transaction{})
	return &GormStore{db}
}

func (s *GormStore) Transactions(address string) ([]Transaction, error) {
	return []Transaction{}, nil
}

func (s *GormStore) Transaction(address, txId string) (Transaction, error) {
	return EmptyTransaction, nil
}

func (s *GormStore) InsertTransaction(*Transaction) error {
	return nil
}

func (s *GormStore) UpdateTransaction(*Transaction) error {
	return nil
}
