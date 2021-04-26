package store

import "github.com/onflow/flow-go-sdk"

type DataStore interface {
	AccountStore
}

type AccountStore interface {
	Account(address flow.Address) (Account, error)
	InsertAccount(a Account) error
	DeleteAccount(address flow.Address) error
	AccountKey(address flow.Address) (AccountKey, error)
	InsertAccountKey(k AccountKey) error
	DeleteAccountKey(address flow.Address) error
}
