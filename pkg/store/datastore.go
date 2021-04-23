package store

import "github.com/google/uuid"

type DataStore interface {
	AccountStore
}

type AccountStore interface {
	Account(id uuid.UUID) (Account, error)
	Accounts() ([]Account, error)
	CreateAccount(a *Account) error
	DeleteAccount(id uuid.UUID) error
}
