package transactions

import "github.com/eqlabs/flow-wallet-service/datastore"

// Store manages data regarding transactions.
type Store interface {
	Transactions(tType Type, address string, opt datastore.ListOptions) ([]Transaction, error)
	Transaction(tType Type, address, txId string) (Transaction, error)
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
