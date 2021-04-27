package store

const (
	DB_TYPE_MEMORY     = "memory"
	DB_TYPE_POSTGRESQL = "psql"
	DB_TYPE_MYSQL      = "mysql"
	DB_TYPE_SQLITE     = "sqlite"
)

type DataStore interface {
	AccountStore
}

type AccountStore interface {
	Accounts() ([]Account, error)
	Account(address string) (Account, error)
	InsertAccount(a Account) error
	DeleteAccount(address string) error
	AccountKeys() ([]AccountKey, error)
	AccountKey(address string) (AccountKey, error)
	InsertAccountKey(k AccountKey) error
	DeleteAccountKey(address string) error
}
