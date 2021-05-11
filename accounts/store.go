package accounts

// Store manages data regarding accounts.
type Store interface {
	// List all accounts.
	Accounts() ([]Account, error)
	// Get account details.
	Account(address string) (Account, error)
	// Insert a new account.
	InsertAccount(a *Account) error
}
