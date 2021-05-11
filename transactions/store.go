package transactions

// TransactionStore manages data regarding transactions.
type TransactionStore interface {
	Transactions(address string) ([]Transaction, error)
	Transaction(address, txId string) (Transaction, error)
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
