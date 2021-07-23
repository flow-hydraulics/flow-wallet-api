package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-api/jobs"
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

func (s *Jobs) List() http.Handler {
	return http.HandlerFunc(s.ListFunc)
}

func (s *Jobs) Details() http.Handler {
	return http.HandlerFunc(s.DetailsFunc)
}
