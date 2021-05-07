package jobs

import (
	"github.com/google/uuid"
)

// JobStore manages data regarding jobs.
type JobStore interface {
	Job(id uuid.UUID) (job Job, err error)
	InsertJob(job *Job) error
	UpdateJob(job *Job) error
}
