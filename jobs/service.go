package jobs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/google/uuid"
)

// Service defines the API for job HTTP handlers.
type Service struct {
	log *log.Logger
	db  Store
}

// NewService initiates a new job service.
func NewService(l *log.Logger, db Store) *Service {
	return &Service{l, db}
}

// Details returns a specific job.
func (s *Service) Details(jobId string) (result Job, err error) {
	id, err := uuid.Parse(jobId)
	if err != nil {
		// Convert error to a 400 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid job id"),
		}
		return
	}

	// Get from datastore
	result, err = s.db.Job(id)
	if err != nil && err.Error() == "record not found" {
		// Convert error to a 404 RequestError
		err = &errors.RequestError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("job not found"),
		}
		return
	}

	return
}
