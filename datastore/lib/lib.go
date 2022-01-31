package lib

import "gorm.io/gorm"

// GormTransaction performs a function on a gorm database transaction instance
// when using something else than sqlite as the dialector (mysql or psql).
// When using sqlite it will fallback to regular gorm database instance.
func GormTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	isSqlite := db.Config.Dialector.Name() == "sqlite"

	var tx *gorm.DB

	if !isSqlite {
		tx = db.Begin()
	} else {
		tx = db
	}

	if err := fn(tx); err != nil {
		if !isSqlite {
			tx.Rollback()
		}
		return err
	}

	if !isSqlite {
		tx.Commit()
	}

	return nil
}
