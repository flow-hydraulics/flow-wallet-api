// m20211202 handles IdempotencyStoreGormItem migration
// NOTE: IdempotencyStoreGormItems are used to store idempotency keys
// when idempotency middleware is enabled & configured to use the shared sql database
package m20211202

import (
	"time"

	"gorm.io/gorm"
)

const ID = "20211202"

type IdempotencyStoreGormItem struct {
	Key        string    `gorm:"column:key;primary_key"`
	ExpiryDate time.Time `gorm:"column:expiry_date"`
}

func (IdempotencyStoreGormItem) TableName() string {
	return "idempotency_keys"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&IdempotencyStoreGormItem{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable(&IdempotencyStoreGormItem{}); err != nil {
		return err
	}

	return nil
}
