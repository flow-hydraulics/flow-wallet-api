// Package tokens provides functions for token handling in Flow blockhain.
// https://docs.onflow.org/core-contracts
package tokens

import (
	"time"

	"github.com/eqlabs/flow-wallet-service/transactions"
	"gorm.io/gorm"
)

// FungibleTokenTransfer is used for database interfacing
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

// FungibleTokenTransferBase is used for JSON interfacing
type FungibleTokenTransferBase struct {
	TransactionId string    `json:"transactionId"`
	Amount        string    `json:"amount"`
	TokenName     string    `json:"token"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// FungibleTokenWithdrawal is used for JSON interfacing
type FungibleTokenWithdrawal struct {
	FungibleTokenTransferBase
	RecipientAddress string `json:"recipient"`
}

// FungibleTokenDeposit is used for JSON interfacing
type FungibleTokenDeposit struct {
	FungibleTokenTransferBase
	PayerAddress string `json:"sender"`
}

func baseFromTransfer(t *FungibleTokenTransfer) FungibleTokenTransferBase {
	return FungibleTokenTransferBase{
		TransactionId: t.TransactionId,
		Amount:        t.Amount,
		TokenName:     t.TokenName,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func (t *FungibleTokenTransfer) Withdrawal() FungibleTokenWithdrawal {
	return FungibleTokenWithdrawal{
		baseFromTransfer(t),
		t.RecipientAddress,
	}
}

func (t *FungibleTokenTransfer) Deposit() FungibleTokenDeposit {
	return FungibleTokenDeposit{
		baseFromTransfer(t),
		t.Transaction.PayerAddress,
	}
}
