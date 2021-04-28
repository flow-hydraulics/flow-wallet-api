package data

const (
	DB_TYPE_POSTGRESQL = "psql"
	DB_TYPE_MYSQL      = "mysql"
	DB_TYPE_SQLITE     = "sqlite"
)

type Store interface {
	AccountStore
}

type AccountStore interface {
	Accounts() ([]Account, error)
	InsertAccount(a Account) error
	Account(address string) (Account, error)
	AccountKeys() ([]AccountKey, error)
	InsertAccountKey(k AccountKey) error
	AccountKey(address string) (AccountKey, error)
}
