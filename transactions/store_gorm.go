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

func (s *GormStore) Transactions(address string) (transactions []Transaction, err error) {
	err = s.db.Find(&transactions).Error
	return
}

func (s *GormStore) Transaction(address, txId string) (transaction Transaction, err error) {
	err = s.db.Where(&Transaction{AccountAddress: address, TransactionId: txId}).First(&transaction).Error
	return
}

func (s *GormStore) InsertTransaction(transaction *Transaction) error {
	return s.db.Create(transaction).Error
}

func (s *GormStore) UpdateTransaction(transaction *Transaction) error {
	return s.db.Save(transaction).Error
}
