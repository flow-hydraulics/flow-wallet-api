package tokens

import "github.com/eqlabs/flow-wallet-api/templates"

// Store manages data regarding tokens.
type Store interface {
	// List an accounts enabled tokens
	AccountTokens(address string, tType *templates.TokenType) ([]AccountToken, error)

	// Enable a token for an account
	InsertAccountToken(at *AccountToken) error

	InsertFungibleTokenTransfer(*TokenTransfer) error
	FungibleTokenWithdrawals(address, tokenName string) ([]*TokenTransfer, error)
	FungibleTokenWithdrawal(address, tokenName, transactionId string) (*TokenTransfer, error)
	FungibleTokenDeposits(address, tokenName string) ([]*TokenTransfer, error)
	FungibleTokenDeposit(address, tokenName, transactionId string) (*TokenTransfer, error)
}
