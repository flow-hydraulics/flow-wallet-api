package tokens

import "github.com/eqlabs/flow-wallet-api/templates"

// Store manages data regarding tokens.
type Store interface {
	// List an accounts enabled tokens
	AccountTokens(address string, tType *templates.TokenType) ([]AccountToken, error)

	// Enable a token for an account
	InsertAccountToken(at *AccountToken) error

	InsertFungibleTokenTransfer(*FungibleTokenTransfer) error
	FungibleTokenWithdrawals(address, tokenName string) ([]*FungibleTokenTransfer, error)
	FungibleTokenWithdrawal(address, tokenName, transactionId string) (*FungibleTokenTransfer, error)
	FungibleTokenDeposits(address, tokenName string) ([]*FungibleTokenTransfer, error)
	FungibleTokenDeposit(address, tokenName, transactionId string) (*FungibleTokenTransfer, error)
}
