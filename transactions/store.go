package transactions

import "github.com/eqlabs/flow-wallet-service/datastore"

// Store manages data regarding transactions.
type Store interface {
	Transactions(address string, opt datastore.ListOptions) ([]Transaction, error)
	Transaction(address, txId string) (Transaction, error)
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
