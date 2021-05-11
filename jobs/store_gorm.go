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

func (s *GormStore) Job(id uuid.UUID) (job Job, err error) {
	err = s.db.First(&job, "id = ?", id).Error
	return
}

func (s *GormStore) InsertJob(job *Job) error {
	return s.db.Create(job).Error
}

func (s *GormStore) UpdateJob(job *Job) error {
	return s.db.Save(job).Error
}
