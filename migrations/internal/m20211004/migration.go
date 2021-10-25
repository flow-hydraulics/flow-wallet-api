package m20211004

import (
	"time"

	"gorm.io/gorm"
)

const ID = "20211004"

type Transaction struct {
	TransactionId   string `gorm:"column:transaction_id;primaryKey"`
	TransactionType int    `gorm:"column:transaction_type;index"`
	Code            TransactionCode
	Arguments       []TransactionArgument
	ProposerAddress string         `gorm:"column:proposer_address;index"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type TransactionCode struct {
	TransactionID string `gorm:"column:transaction_id;primaryKey"`
	Code          string `gorm:"column:code"`
}

type TransactionArgument struct {
	TransactionID string `gorm:"column:transaction_id:primaryKey"`
	Index         int    `gorm:"column:index:primaryKey"`
	Type          string `gorm:column:type"`
	Value         string `gorm:column:value"`
}

func (TransactionArgument) TableName() string {
	return "transaction_arguments"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&TransactionCode{}); err != nil {
		return err
	}

	if err := tx.AutoMigrate(&TransactionArgument{}); err != nil {
		return err
	}

	if err := tx.AutoMigrate(&Transaction{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable(&TransactionCode{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&TransactionArgument{}); err != nil {
		return err
	}

	return nil
}
