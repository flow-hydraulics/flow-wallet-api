package gorm

import (
	"github.com/eqlabs/flow-nft-wallet-service/data"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
}

func NewStore(dialector gorm.Dialector) (*Store, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return &Store{}, err
	}
	return &Store{
		AccountStore: newAccountStore(db),
	}, nil
}
