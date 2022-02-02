package system

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetSettings() (*Settings, error)
	SaveSettings(settings *Settings) error
	// Pauses the system for a configured duration
	Pause() error
	// Will immediately revoke the pause timer
	Resume() error
	IsHalted() (bool, error)
}

type ServiceImpl struct {
	store         Store
	pauseDuration time.Duration
}

const defaultPauseDuration = time.Minute

func NewService(store Store, opts ...ServiceOption) Service {
	svc := &ServiceImpl{
		store:         store,
		pauseDuration: defaultPauseDuration,
	}

	// Go through options
	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func (svc *ServiceImpl) GetSettings() (*Settings, error) {
	return svc.store.GetSettings()
}

func (svc *ServiceImpl) SaveSettings(settings *Settings) error {
	if settings.ID == 0 {
		return fmt.Errorf("settings object has no ID, get an existing settings first and alter it")
	}
	log.WithFields(log.Fields{"settings": settings}).Trace("Save system settings")
	return svc.store.SaveSettings(settings)
}

func (svc *ServiceImpl) Pause() error {
	log.Trace("Pause system")
	settings, err := svc.GetSettings()
	if err != nil {
		return err
	}
	settings.PausedSince = sql.NullTime{Time: time.Now(), Valid: true}
	return svc.SaveSettings(settings)
}

func (svc *ServiceImpl) Resume() error {
	log.Trace("Resume system")
	settings, err := svc.GetSettings()
	if err != nil {
		return err
	}
	settings.PausedSince = sql.NullTime{Time: time.Now(), Valid: false}
	return svc.SaveSettings(settings)
}

func (svc *ServiceImpl) IsHalted() (bool, error) {
	s, err := svc.GetSettings()
	if err != nil {
		return false, err
	}
	return s.IsMaintenanceMode() || s.IsPaused(svc.pauseDuration), nil
}
