package chain_events

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&ListenerStatus{})
	return &GormStore{db}
}

func (s *GormStore) GetListenerStatus() (t *ListenerStatus, err error) {
	err = s.db.FirstOrCreate(&t).Error
	return
}

func (s *GormStore) UpdateListenerStatus(t *ListenerStatus) error {
	return s.db.Save(&t).Error
}
