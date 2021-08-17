package transactions

import "github.com/flow-hydraulics/flow-wallet-api/datastore"

// Store manages data regarding transactions.
type Store interface {
	Transactions(tType Type, address string, opt datastore.ListOptions) ([]Transaction, error)
	Transaction(tType Type, address, txId string) (Transaction, error)
	GetOrCreateTransaction(txId string) *Transaction
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
