package m20211118

import (
	"time"

	"gorm.io/gorm"
)

const ID = "20211118"

type Storable struct {
	ID             int            `json:"-" gorm:"primaryKey"`
	AccountAddress string         `json:"-" gorm:"index"`
	Index          int            `json:"index" gorm:"index"`
	Type           string         `json:"type"`
	Value          []byte         `json:"-"`
	PublicKey      string         `json:"publicKey"`
	SignAlgo       string         `json:"signAlgo"`
	HashAlgo       string         `json:"hashAlgo"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Storable) TableName() string {
	return "storable_keys"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Storable{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&Storable{}, "PublicKey"); err != nil {
		return err
	}

	return nil
}
