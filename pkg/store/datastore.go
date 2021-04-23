package store

import "github.com/google/uuid"

type DataStore interface {
	Account(id uuid.UUID) (Account, error)
	Accounts() ([]Account, error)
	Transaction(id uuid.UUID) (Transaction, error)
	Transactions() ([]Transaction, error)
}
