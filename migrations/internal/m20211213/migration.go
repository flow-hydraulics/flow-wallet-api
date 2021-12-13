package m20211213

import (
	"database/sql"

	"gorm.io/gorm"
)

const ID = "m20211213"

type Settings struct {
	gorm.Model
	MaintenanceMode bool         `gorm:"column:maintenance_mode;default:false"`
	PausedSince     sql.NullTime `gorm:"column:paused_since"`
}

func (Settings) TableName() string {
	return "system_settings"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Settings{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&Settings{}, "paused_since"); err != nil {
		return err
	}

	return nil
}
