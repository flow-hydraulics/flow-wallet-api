package transactions

// Store manages data regarding transactions.
type Store interface {
	Transactions(address string) ([]Transaction, error)
	Transaction(address, txId string) (Transaction, error)
	InsertTransaction(*Transaction) error
	UpdateTransaction(*Transaction) error
}
