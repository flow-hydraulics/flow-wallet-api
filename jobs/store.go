package jobs

import (
	"github.com/google/uuid"
)

// Store manages data regarding jobs.
type Store interface {
	Job(id uuid.UUID) (Job, error)
	InsertJob(*Job) error
	UpdateJob(*Job) error
}
