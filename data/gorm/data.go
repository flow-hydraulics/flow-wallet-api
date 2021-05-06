// Package gorm manages data storage in sqlite, postgressql and mysql.
package gorm

import (
	"log"

	"github.com/eqlabs/flow-wallet-service/data"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
}

// NewStore initiates a new store.
func NewStore(l *log.Logger) (result *Store, err error) {
	cfg := ParseConfig()

	db, err := gorm.Open(cfg.Dialector, cfg.Options)
	if err != nil {
		return
	}

	accountStore, err := NewAccountStore(l, db)
	if err != nil {
		return
	}

	result = &Store{
		AccountStore: accountStore,
	}

	return
}
