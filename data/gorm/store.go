package gorm

import (
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/data"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
}

func NewStore(l *log.Logger) (result *Store, err error) {
	db, err := gorm.Open(cfg.Dialector, &gorm.Config{})
	if err != nil {
		return
	}

	result = &Store{
		AccountStore: newAccountStore(l, db),
	}

	return
}
