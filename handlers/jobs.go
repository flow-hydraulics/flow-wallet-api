package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/gorilla/mux"
)

// Jobs is a HTTP server for jobs.
// It provides details API.
// It uses jobs service to interface with data.
type Jobs struct {
	log     *log.Logger
	service *jobs.Service
}

// NewJobs initiates a new jobs server.
func NewJobs(l *log.Logger, service *jobs.Service) *Jobs {
	return &Jobs{l, service}
}

// Details returns details regarding a job.
// It reads the job id for the wanted job from URL.
// Job service is responsible for validating the job id.
func (s *Jobs) Details(rw http.ResponseWriter, r *http.Request) {
	s.log.Println("Job details")
	vars := mux.Vars(r)
	res, err := s.service.Details(vars["jobId"])
	if err != nil {
		handleError(err, s.log, rw)
		return
	}
	handleJsonResponse(rw, http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
