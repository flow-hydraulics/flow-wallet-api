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
	NftID     uint64 `json:"nftId,omitempty"`
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
	NftID            uint64                   `json:"nftId"`
	TokenName        string                   `json:"token"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt           `json:"-" gorm:"index"`
}

// TokenTransferBase is used for JSON interfacing
type TokenTransferBase struct {
	TransactionId string    `json:"transactionId"`
	FtAmount      string    `json:"amount"`
	NftID         uint64    `json:"nftId"`
	TokenName     string    `json:"token"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TokenWithdrawal is used for JSON interfacing
type TokenWithdrawal struct {
	TokenTransferBase
	RecipientAddress string `json:"recipient"`
}

// TokenDeposit is used for JSON interfacing
type TokenDeposit struct {
	TokenTransferBase
	PayerAddress string `json:"sender"`
}

func baseFromTransfer(t *TokenTransfer) TokenTransferBase {
	return TokenTransferBase{
		TransactionId: t.TransactionId,
		FtAmount:      t.FtAmount,
		NftID:         t.NftID,
		TokenName:     t.TokenName,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func (t *TokenTransfer) Withdrawal() TokenWithdrawal {
	return TokenWithdrawal{
		baseFromTransfer(t),
		t.RecipientAddress,
	}
}

func (t *TokenTransfer) Deposit() TokenDeposit {
	return TokenDeposit{
		baseFromTransfer(t),
		t.Transaction.PayerAddress,
	}
}
