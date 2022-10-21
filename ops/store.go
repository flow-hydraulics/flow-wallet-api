package ops

import (
	"github.com/flow-hydraulics/flow-wallet-api/accounts"
)

// Store defines what ops needs from the database
type Store interface {
	ListAccountsWithMissingVault(tokenName string) (*[]accounts.Account, error)
}
