package gorm

import (
	"log"

	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobStore struct {
	l  *log.Logger
	db *gorm.DB
}

func NewJobStore(l *log.Logger, db *gorm.DB) (*JobStore, error) {
	err := db.AutoMigrate(&jobs.Job{})
	if err != nil {
		return &JobStore{}, err
	}
	return &JobStore{l, db}, nil
}

func (s *JobStore) InsertJob(job *jobs.Job) error {
	return s.db.Create(job).Error
}

func (s *JobStore) UpdateJob(job *jobs.Job) error {
	return s.db.Save(job).Error
}

func (s *JobStore) Job(id uuid.UUID) (job jobs.Job, err error) {
	err = s.db.First(&job, "id = ?", id).Error
	return
}
