package transactions

import (
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
)

// Store manages data regarding transactions.
type Store interface {
	Transactions(opt datastore.ListOptions) ([]Transaction, error)
	Transaction(txId string) (Transaction, error)
	TransactionsForAccount(tType Type, address string, opt datastore.ListOptions) ([]Transaction, error)
	TransactionForAccount(tType Type, address, txId string) (Transaction, error)
	GetOrCreateTransaction(txId string) *Transaction
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
