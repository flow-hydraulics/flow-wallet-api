package jobs

import (
	"github.com/eqlabs/flow-wallet-api/datastore"
	"github.com/google/uuid"
)

// Store manages data regarding jobs.
type Store interface {
	Jobs(datastore.ListOptions) ([]Job, error)
	Job(id uuid.UUID) (Job, error)
	InsertJob(*Job) error
	UpdateJob(*Job) error
}
