package jobs

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Job{})
	return &GormStore{db}
}

func (s *GormStore) Job(id uuid.UUID) (j Job, err error) {
	err = s.db.First(&j, "id = ?", id).Error
	return
}

func (s *GormStore) InsertJob(j *Job) error {
	return s.db.Create(j).Error
}

func (s *GormStore) UpdateJob(j *Job) error {
	return s.db.Save(j).Error
}
