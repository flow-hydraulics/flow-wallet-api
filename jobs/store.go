package jobs

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/google/uuid"
)

// Store manages data regarding jobs.
type Store interface {
	Jobs(datastore.ListOptions) ([]Job, error)
	Job(id uuid.UUID) (Job, error)
	InsertJob(*Job) error
	UpdateJob(*Job) error
	AcceptJob(j *Job, acceptedGracePeriod time.Duration) error
	SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod time.Duration, o datastore.ListOptions) ([]Job, error)
	Status() ([]StatusQuery, error)
}

type StatusQuery struct {
	State State
	Count int
}
