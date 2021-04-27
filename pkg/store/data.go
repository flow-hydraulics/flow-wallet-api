package store

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	Address   string         `json:"address" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type AccountKey struct {
	AccountAddress string         `json:"address" gorm:"primaryKey"`
	Index          int            `json:"index"`
	Type           string         `json:"type"`  // local, google_kms
	Value          string         `json:"value"` // local: private key, google_kms: resource id
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
