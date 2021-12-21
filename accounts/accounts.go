// Package account provides functions for account management on Flow blockhain.
package accounts

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"gorm.io/gorm"
)

type AccountType string

const AccountTypeCustodial = "custodial"
const AccountTypeNonCustodial = "non-custodial"

// Account struct represents a storable account.
type Account struct {
	Address   string          `json:"address" gorm:"primaryKey"`
	Keys      []keys.Storable `json:"keys" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Type      AccountType     `json:"type" gorm:"default:custodial"`
	CreatedAt time.Time       `json:"createdAt" `
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index"`
}
