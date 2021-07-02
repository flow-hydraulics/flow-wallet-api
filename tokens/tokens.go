// Package tokens provides functions for token handling for a hosted account on Flow Wallet API.

package tokens

import (
	"time"

	"github.com/eqlabs/flow-wallet-api/transactions"
	"gorm.io/gorm"
)

type Details struct {
	TokenName string `json:"name"`
	Address   string `json:"address,omitempty"`
	Balance   string `json:"balance,omitempty"`
}

type WithdrawalRequest struct {
	TokenName string `json:"-"`
	Recipient string `json:"recipient"`
	FtAmount  string `json:"amount,omitempty"`
	NftID     string `json:"id,omitempty"`
}

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
