// Package tokens provides functions for token handling for a hosted account on Flow Wallet API.

package tokens

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"gorm.io/gorm"
)

type Details struct {
	TokenName string   `json:"name"`
	Address   string   `json:"address,omitempty"`
	Balance   *Balance `json:"balance,omitempty"`
}

type WithdrawalRequest struct {
	TokenName string `json:"tokenName,omitempty"`
	Recipient string `json:"recipient"`
	FtAmount  string `json:"amount,omitempty"`
	NftID     uint64 `json:"nftId,omitempty"`
}

// AccountToken represents a token that is enabled on an account.
type AccountToken struct {
	ID             uint64              `json:"-" gorm:"column:id;primaryKey"`
	AccountAddress string              `json:"-" gorm:"column:account_address;uniqueIndex:addressname;index;not null"`
	Account        accounts.Account    `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenName      string              `json:"name" gorm:"column:token_name;uniqueIndex:addressname;index;not null"`
	TokenAddress   string              `json:"address" gorm:"column:token_address;uniqueIndex:addressname;index;not null"`
	TokenType      templates.TokenType `json:"-" gorm:"column:token_type"`
	CreatedAt      time.Time           `json:"-" gorm:"column:created_at"`
	UpdatedAt      time.Time           `json:"-" gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt      `json:"-" gorm:"column:deleted_at;index"`
}

func (AccountToken) TableName() string {
	return "account_tokens"
}

// TokenTransfer is used for database interfacing
type TokenTransfer struct {
	ID               uint64                   `gorm:"column:id;primaryKey"`
	TransactionId    string                   `gorm:"column:transaction_id"` // TODO (latenssi): should propably be unique over this column
	Transaction      transactions.Transaction `gorm:"foreignKey:TransactionId;references:TransactionId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RecipientAddress string                   `gorm:"column:recipient_address;index"`
	SenderAddress    string                   `gorm:"column:sender_address;index"`
	FtAmount         string                   `gorm:"column:ft_amount"`
	NftID            uint64                   `gorm:"column:nft_id"`
	TokenName        string                   `gorm:"column:token_name"`
	CreatedAt        time.Time                `gorm:"column:created_at"`
	UpdatedAt        time.Time                `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt           `gorm:"column:deleted_at;index"`
}

func (TokenTransfer) TableName() string {
	return "token_transfers"
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
	SenderAddress string `json:"sender"`
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
		t.SenderAddress,
	}
}
