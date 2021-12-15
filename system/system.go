package system

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	MaintenanceMode bool         `gorm:"column:maintenance_mode;default:false"`
	PausedSince     sql.NullTime `gorm:"column:paused_since"`
}

func (s *Settings) String() string {
	return fmt.Sprintf("MaintenanceMode: %t", s.MaintenanceMode)
}

func (Settings) TableName() string {
	return "system_settings"
}

// Convert to JSON version
func (s *Settings) ToJSON() SettingsJSON {
	return SettingsJSON{
		MaintenanceMode: s.MaintenanceMode,
	}
}

func (s *Settings) IsMaintenanceMode() bool {
	return s.MaintenanceMode
}

func (s *Settings) IsPaused(pauseDuration time.Duration) bool {
	return s.PausedSince.Valid && s.PausedSince.Time.After(time.Now().Add(-pauseDuration))
}

// Update fields according to JSON version
func (s *Settings) FromJSON(j SettingsJSON) {
	s.MaintenanceMode = j.MaintenanceMode
}

type SettingsJSON struct {
	MaintenanceMode bool `json:"maintenanceMode"`
}
