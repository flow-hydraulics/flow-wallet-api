package gorm

import "gorm.io/gorm"

func New() (*gorm.DB, error) {
	cfg := ParseConfig()
	db, err := gorm.Open(cfg.Dialector, cfg.Options)
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
	err = sqlDB.Close()
	if err != nil {
		panic("unable to close database")
	}
}
