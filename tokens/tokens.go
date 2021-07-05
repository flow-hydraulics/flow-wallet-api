// Package tokens provides functions for token handling for a hosted account on Flow Wallet API.

package tokens

import (
	"time"

	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/eqlabs/flow-wallet-api/transactions"
	"github.com/onflow/cadence"
	"gorm.io/gorm"
)

type Details struct {
	TokenName string        `json:"name"`
	Address   string        `json:"address,omitempty"`
	Balance   cadence.Value `json:"balance,omitempty"`
}

type WithdrawalRequest struct {
	TokenName string `json:"-"`
	Recipient string `json:"recipient"`
	FtAmount  string `json:"amount,omitempty"`
	NftID     uint64 `json:"id,omitempty"`
}

// AccountToken represents a token that is enabled on an account.
type AccountToken struct {
	ID             uint64              `json:"-" gorm:"primaryKey"`
	AccountAddress string              `json:"-" gorm:"uniqueIndex:addressname;index;not null"`
	Account        accounts.Account    `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenName      string              `json:"name" gorm:"uniqueIndex:addressname;index;not null"`
	TokenAddress   string              `json:"address" gorm:"uniqueIndex:addressname;index;not null"`
	TokenType      templates.TokenType `json:"-"`
	CreatedAt      time.Time           `json:"-"`
	UpdatedAt      time.Time           `json:"-"`
	DeletedAt      gorm.DeletedAt      `json:"-" gorm:"index"`
}

// TokenTransfer is used for database interfacing
type TokenTransfer struct {
	ID               uint64                   `json:"-" gorm:"primaryKey"`
	TransactionId    string                   `json:"transactionId"`
	Transaction      transactions.Transaction `json:"-" gorm:"foreignKey:TransactionId;references:TransactionId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RecipientAddress string                   `json:"recipient" gorm:"index"`
	FtAmount         string                   `json:"amount"`
	NftID            uint64                   `json:"id"`
	TokenName        string                   `json:"token"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt           `json:"-" gorm:"index"`
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

func baseFromTransfer(t *TokenTransfer) FungibleTokenTransferBase {
	return FungibleTokenTransferBase{
		TransactionId: t.TransactionId,
		Amount:        t.FtAmount,
		TokenName:     t.TokenName,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func (t *TokenTransfer) Withdrawal() FungibleTokenWithdrawal {
	return FungibleTokenWithdrawal{
		baseFromTransfer(t),
		t.RecipientAddress,
	}
}

func (t *TokenTransfer) Deposit() FungibleTokenDeposit {
	return FungibleTokenDeposit{
		baseFromTransfer(t),
		t.Transaction.PayerAddress,
	}
}
