package tokens

import "github.com/flow-hydraulics/flow-wallet-api/templates"

// Store manages data regarding tokens.
type Store interface {
	// List an accounts enabled tokens
	AccountTokens(address string, tokenType templates.TokenType) ([]AccountToken, error)

	// Enable a token for an account
	InsertAccountToken(at *AccountToken) error

	InsertTokenTransfer(*TokenTransfer) error
	TokenWithdrawals(address string, token *templates.Token) ([]*TokenTransfer, error)
	TokenWithdrawal(address, transactionId string, token *templates.Token) (*TokenTransfer, error)
	TokenDeposits(address string, token *templates.Token) ([]*TokenTransfer, error)
	TokenDeposit(address, transactionId string, token *templates.Token) (*TokenTransfer, error)
}
