// Package gorm manages data storage in sqlite, postgressql and mysql.
package gorm

import (
	"log"

	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/eqlabs/flow-wallet-service/jobs"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
	jobs.JobStore
}

// NewStore initiates a new gorm data store.
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

	jobStore, err := NewJobStore(l, db)
	if err != nil {
		return
	}

	result = &Store{
		AccountStore: accountStore,
		JobStore:     jobStore,
	}

	return
}
