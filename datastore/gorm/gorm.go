package gorm

import (
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/migrations"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbTypePostgresql = "psql"
	dbTypeMysql      = "mysql"
	dbTypeSqlite     = "sqlite"
)

func New(cfg *configs.Config) (*gorm.DB, error) {
	// TODO(latenssi): safeguard against nil config?

	var dialector gorm.Dialector
	switch cfg.DatabaseType {
	default:
		panic(fmt.Sprintf("database type '%s' not supported", cfg.DatabaseType))
	case dbTypePostgresql:
		dialector = postgres.Open(cfg.DatabaseDSN)
	case dbTypeMysql:
		dialector = mysql.Open(cfg.DatabaseDSN)
	case dbTypeSqlite:
		dialector = sqlite.Open(cfg.DatabaseDSN)
	}

	options := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(dialector, options)
	if err != nil {
		return &gorm.DB{}, err
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.List())
	if cfg.DatabaseVersion == "" {
		err = m.Migrate()
	} else {
		err = m.MigrateTo(cfg.DatabaseVersion)
		if err != nil {
			return &gorm.DB{}, err
		}

		err = m.RollbackTo(cfg.DatabaseVersion)
	}
	if err != nil {
		return &gorm.DB{}, err
	}

	return db, nil
}

func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic("unable to close database")
	}

	if err := sqlDB.Close(); err != nil {
		panic("unable to close database")
	}
}
