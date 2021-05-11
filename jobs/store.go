package jobs

import (
	"github.com/google/uuid"
)

// Store manages data regarding jobs.
type Store interface {
	Job(id uuid.UUID) (job Job, err error)
	InsertJob(job *Job) error
	UpdateJob(job *Job) error
}
