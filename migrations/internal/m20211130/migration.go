package m20211130

import (
	"gorm.io/gorm"
)

// Note: there is an 'm' here, it is a typo but it should not be removed
const ID = "m20211130"

type Settings struct {
	gorm.Model
	MaintenanceMode bool `gorm:"column:maintenance_mode;default:false"`
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
	if err := tx.Migrator().DropTable(&Settings{}); err != nil {
		return err
	}

	return nil
}
