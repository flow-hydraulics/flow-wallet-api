package system

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &GormStore{db}
}

func (s *GormStore) GetSettings() (*Settings, error) {
	settings := &Settings{}
	return settings, s.db.FirstOrCreate(settings).Error
}

func (s *GormStore) SaveSettings(settings *Settings) error {
	return s.db.Save(&settings).Error
}
