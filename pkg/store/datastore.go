package store

import "github.com/onflow/flow-go-sdk"

type DataStore interface {
	AccountStore
}

type AccountStore interface {
	Accounts() ([]Account, error)
	Account(address flow.Address) (Account, error)
	InsertAccount(a Account) error
	DeleteAccount(address flow.Address) error
	AccountKeys() ([]AccountKey, error)
	AccountKey(address flow.Address) (AccountKey, error)
	InsertAccountKey(k AccountKey) error
	DeleteAccountKey(address flow.Address) error
}
