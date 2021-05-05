package gorm

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DB_TYPE_POSTGRESQL = "psql"
	DB_TYPE_MYSQL      = "mysql"
	DB_TYPE_SQLITE     = "sqlite"
)

type Config struct {
	DatabaseDSN  string `env:"DB_DSN" envDefault:"wallet.db"`
	DatabaseType string `env:"DB_TYPE" envDefault:"sqlite"`
	Dialector    gorm.Dialector
	Options      *gorm.Config
}

func ParseConfig() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	var d gorm.Dialector
	switch cfg.DatabaseType {
	case DB_TYPE_POSTGRESQL:
		d = postgres.Open(cfg.DatabaseDSN)
	case DB_TYPE_MYSQL:
		d = mysql.Open(cfg.DatabaseDSN)
	case DB_TYPE_SQLITE:
		d = sqlite.Open(cfg.DatabaseDSN)
	default:
		panic(fmt.Sprintf("database type '%s' not supported", cfg.DatabaseType))
	}

	cfg.Dialector = d
	cfg.Options = &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	return
}
