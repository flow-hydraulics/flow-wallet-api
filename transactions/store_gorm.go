package transactions

import (
	"github.com/eqlabs/flow-wallet-service/datastore"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Transaction{})
	return &GormStore{db}
}

func (s *GormStore) Transactions(address string, o datastore.ListOptions) (tt []Transaction, err error) {
	err = s.db.
		Where(&Transaction{PayerAddress: address}).
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&tt).Error
	return
}

func (s *GormStore) Transaction(address, txId string) (t Transaction, err error) {
	err = s.db.Where(&Transaction{PayerAddress: address, TransactionId: txId}).First(&t).Error
	return
}

func (s *GormStore) InsertTransaction(t *Transaction) error {
	return s.db.Create(t).Error
}

func (s *GormStore) UpdateTransaction(t *Transaction) error {
	return s.db.Save(t).Error
}
