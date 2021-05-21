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
	DATABASE_TYPE_POSTGRESQL = "psql"
	DATABASE_TYPE_MYSQL      = "mysql"
	DATABASE_TYPE_SQLITE     = "sqlite"
)

// Config struct for gorm data store.
type Config struct {
	DatabaseDSN  string `env:"DATABASE_DSN" envDefault:"wallet.db"`
	DatabaseType string `env:"DATABASE_TYPE" envDefault:"sqlite"`
	Dialector    gorm.Dialector
	Options      *gorm.Config
}

// ParseConfig parses environment variables to a valid Config.
func ParseConfig() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	var d gorm.Dialector
	switch cfg.DatabaseType {
	case DATABASE_TYPE_POSTGRESQL:
		d = postgres.Open(cfg.DatabaseDSN)
	case DATABASE_TYPE_MYSQL:
		d = mysql.Open(cfg.DatabaseDSN)
	case DATABASE_TYPE_SQLITE:
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
