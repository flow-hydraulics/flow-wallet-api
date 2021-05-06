// Package data manages data storage.
package data

import (
	"time"

	"gorm.io/gorm"
)

// Store interface groups different types of stores together.
type Store interface {
	AccountStore
}

// AccountStore manages data regarding accounts.
type AccountStore interface {
	// List all accounts.
	Accounts() ([]Account, error)
	// Insert a new account.
	InsertAccount(a Account) error
	// Get account details.
	Account(address string) (Account, error)
	// Get the account key corresponding to address and key index.
	AccountKey(address string, index int) (Key, error)
}

// Account struct represents a storable account.
type Account struct {
	Address   string         `json:"address" gorm:"primaryKey"`
	Keys      []Key          `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt time.Time      `json:"createdAt" `
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Key struct represents a storable account key.
// Key.Value will either be a byte representation of
// the actual private key when using local key management
// or a resource id when using a remote key management system (e.g. Google KMS).
// Store package is not responsible for encryption/decryption of the Key.Value;
// that is handled by the "keys" package.
type Key struct {
	ID             int            `json:"-" gorm:"primaryKey"`
	AccountAddress string         `json:"-" gorm:"index"`
	Index          int            `json:"index" gorm:"index"`
	Type           string         `json:"type"`
	Value          []byte         `json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
