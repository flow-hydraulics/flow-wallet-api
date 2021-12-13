package system

import (
	"database/sql"
	"fmt"
	"time"

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
	log.WithFields(log.Fields{"settings": settings}).Trace("Save system settings")
	return svc.store.SaveSettings(settings)
}

func (svc *Service) Pause() error {
	log.Trace("Pause system")
	settings, err := svc.GetSettings()
	if err != nil {
		return err
	}
	settings.PausedSince = sql.NullTime{Time: time.Now(), Valid: true}
	return svc.SaveSettings(settings)
}
