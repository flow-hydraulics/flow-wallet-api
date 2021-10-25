package m20211015

import (
	"time"

	"gorm.io/gorm"
)

const ID = "20211015"

type Transaction struct {
	TransactionId   string         `gorm:"column:transaction_id;primaryKey"`
	TransactionType int            `gorm:"column:transaction_type;index"`
	ProposerAddress string         `gorm:"column:proposer_address;index"`
	FlowTransaction []byte         `gorm:"column:flow_transaction;type:bytes"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Transaction{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&Transaction{}, "FlowTransaction"); err != nil {
		return err
	}

	return nil
}
