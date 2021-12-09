package system

import (
	"fmt"

	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	MaintenanceMode bool `gorm:"column:maintenance_mode;default:false"`
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

// Update fields according to JSON version
func (s *Settings) FromJSON(j SettingsJSON) {
	s.MaintenanceMode = j.MaintenanceMode
}

type SettingsJSON struct {
	MaintenanceMode bool `json:"maintenanceMode"`
}
