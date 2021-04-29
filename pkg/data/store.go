package data

import (
	"time"

	"gorm.io/gorm"
)

const (
	DB_TYPE_POSTGRESQL = "psql"
	DB_TYPE_MYSQL      = "mysql"
	DB_TYPE_SQLITE     = "sqlite"
)

type Store interface {
	AccountStore
}

type AccountStore interface {
	Accounts() ([]Account, error)
	InsertAccount(a Account) error
	Account(address string) (Account, error)
	AccountKey(address string, index int) (Key, error)
}

// Storable account
type Account struct {
	Address   string         `json:"address" gorm:"primaryKey"`
	Keys      []Key          `json:"keys" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Storable account key
type Key struct {
	ID             int            `json:"-" gorm:"primaryKey"`
	AccountAddress string         `json:"-" gorm:"index"`
	Index          int            `json:"index" gorm:"index"`
	Type           string         `json:"type"`
	Value          []byte         `json:"-"`
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
