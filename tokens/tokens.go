// Package tokens provides functions for token handling in Flow blockhain.
// https://docs.onflow.org/core-contracts
package tokens

import (
	"time"

	"github.com/eqlabs/flow-wallet-service/transactions"
	"gorm.io/gorm"
)

type FungibleTokenTransfer struct {
	TransactionId    string                   `json:"transactionId"`
	Transaction      transactions.Transaction `json:"-" gorm:"foreignKey:TransactionId;references:TransactionId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RecipientAddress string                   `json:"recipient" gorm:"index"`
	Amount           string                   `json:"amount"`
	TokenName        string                   `json:"token"`
	ID               int                      `json:"-" gorm:"primaryKey"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt           `json:"-" gorm:"index"`
}

type TokenDetails struct {
	Name    string `json:"name"`
	Balance string `json:"balance"`
}
