package jobs

import (
	"fmt"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	List(limit, offset int) (*[]Job, error)
	Details(jobID string) (*Job, error)
}

// ServiceImpl defines the API for job HTTP handlers.
type ServiceImpl struct {
	store Store
}

// NewService initiates a new job service.
func NewService(store Store) Service {
	return &ServiceImpl{store}
}

// List returns all jobs in the datastore.
func (s *ServiceImpl) List(limit, offset int) (*[]Job, error) {
	log.WithFields(log.Fields{"limit": limit, "offset": offset}).Trace("List jobs")

	o := datastore.ParseListOptions(limit, offset)

	jobs, err := s.store.Jobs(o)
	if err != nil {
		return nil, err
	}

	return &jobs, nil
}

// Details returns a specific job.
func (s *ServiceImpl) Details(jobID string) (*Job, error) {
	log.WithFields(log.Fields{"jobID": jobID}).Trace("Job details")

	id, err := uuid.Parse(jobID)
	if err != nil {
		// Convert error to a 400 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid job id"),
		}
		return nil, err
	}

	// Get from datastore
	job, err := s.store.Job(id)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("job not found"),
		}
		return nil, err
	}

	return &job, nil
}
