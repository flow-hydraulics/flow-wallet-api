package gorm

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-nft-wallet-service/data"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	data.AccountStore
}

type Config struct {
	DatabaseDSN  string `env:"DB_DSN" envDefault:"wallet.db"`
	DatabaseType string `env:"DB_TYPE" envDefault:"sqlite"`
}

func NewStore(l *log.Logger) (store *Store, err error) {
	cfg := Config{}
	if err = env.Parse(&cfg); err != nil {
		return
	}

	var dialector gorm.Dialector
	switch cfg.DatabaseType {
	case data.DB_TYPE_POSTGRESQL:
		dialector = postgres.Open(cfg.DatabaseDSN)
	case data.DB_TYPE_MYSQL:
		dialector = mysql.Open(cfg.DatabaseDSN)
	case data.DB_TYPE_SQLITE:
		dialector = sqlite.Open(cfg.DatabaseDSN)
	default:
		err = fmt.Errorf("database type '%s' not supported", cfg.DatabaseType)
		return
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return
	}

	store = &Store{
		AccountStore: newAccountStore(l, db),
	}

	return
}
