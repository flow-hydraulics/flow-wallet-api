package handlers

import (
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	log "github.com/sirupsen/logrus"
)

// Jobs is a HTTP server for jobs.
// It provides details API.
// It uses jobs service to interface with data.
type Jobs struct {
	logger  *log.Entry
	service *jobs.Service
}

// NewJobs initiates a new jobs server.
func NewJobs(l *log.Entry, service *jobs.Service) *Jobs {
	return &Jobs{l, service}
}

func (s *Jobs) List() http.Handler {
	return http.HandlerFunc(s.ListFunc)
}

func (s *Jobs) Details() http.Handler {
	return http.HandlerFunc(s.DetailsFunc)
}
