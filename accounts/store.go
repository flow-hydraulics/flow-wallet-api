package accounts

import (
	"github.com/eqlabs/flow-wallet-api/datastore"
	"github.com/eqlabs/flow-wallet-api/templates"
)

// Store manages data regarding accounts.
type Store interface {
	// List all accounts.
	Accounts(datastore.ListOptions) ([]Account, error)

	// Get account details.
	Account(address string) (Account, error)

	// Insert a new account.
	InsertAccount(a *Account) error

	// List accounts AccountTokens
	AccountTokens(address string, tType *templates.TokenType) ([]AccountToken, error)

	// Insert an AccountToken.
	InsertAccountToken(at *AccountToken) error
}
