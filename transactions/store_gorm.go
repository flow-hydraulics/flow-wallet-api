package transactions

import (
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db}
}

// -- All transactions

func (s *GormStore) Transactions(o datastore.ListOptions) (tt []Transaction, err error) {
	q := &Transaction{}
	err = s.db.
		Where(q).
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&tt).Error
	return
}

func (s *GormStore) Transaction(txId string) (t Transaction, err error) {
	q := &Transaction{TransactionId: txId}
	err = s.db.Where(q).First(&t).Error
	return
}

// -- Transactions for an account

func (s *GormStore) TransactionsForAccount(tType Type, address string, o datastore.ListOptions) (tt []Transaction, err error) {
	q := &Transaction{ProposerAddress: address, TransactionType: tType}
	err = s.db.
		Where(q).
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&tt).Error
	return
}

func (s *GormStore) TransactionForAccount(tType Type, address, txId string) (t Transaction, err error) {
	q := &Transaction{ProposerAddress: address, TransactionType: tType, TransactionId: txId}
	err = s.db.Where(q).First(&t).Error
	return
}

// -- Misc

func (s *GormStore) GetOrCreateTransaction(txId string) (t *Transaction) {
	s.db.
		Where(&Transaction{TransactionId: txId}).
		Attrs(&Transaction{TransactionType: Unknown}).
		FirstOrCreate(&t)
	return t
}

func (s *GormStore) InsertTransaction(t *Transaction) error {
	return s.db.Create(t).Error
}

func (s *GormStore) UpdateTransaction(t *Transaction) error {
	return s.db.Save(t).Error
}
