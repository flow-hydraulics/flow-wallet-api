package system

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store}
}

func (svc *Service) GetSettings() (*Settings, error) {
	return svc.store.GetSettings()
}

func (svc *Service) SaveSettings(settings *Settings) error {
	if settings.ID == 0 {
		return fmt.Errorf("settings object has no ID, get an existing settings first and alter it")
	}
	return svc.store.SaveSettings(settings)
}

func (svc *Service) IsMaintenanceMode() bool {
	settings, err := svc.GetSettings()
	if err != nil {
		log.
			WithFields(log.Fields{
				"error":    err,
				"package":  "system",
				"function": "IsMaintenanceMode",
			}).
			Warn("Error while getting system settings")
	}
	return err == nil && settings.MaintenanceMode
}
