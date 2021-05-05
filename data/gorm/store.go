package gorm

import (
	"log"

	"github.com/eqlabs/flow-wallet-service/data"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
}

func NewStore(l *log.Logger) (result *Store, err error) {
	cfg := ParseConfig()

	db, err := gorm.Open(cfg.Dialector, cfg.Options)
	if err != nil {
		return
	}

	result = &Store{
		AccountStore: newAccountStore(l, db),
	}

	return
}
