package gorm

import (
	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"gorm.io/gorm"
)

type DataStore struct {
	data.AccountStore
}

func NewDataStore(dialector gorm.Dialector) (*DataStore, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return &DataStore{}, nil
	}
	return &DataStore{
		AccountStore: newAccountStore(db),
	}, nil
}
