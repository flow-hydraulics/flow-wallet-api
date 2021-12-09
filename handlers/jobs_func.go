package handlers

import (
	"net/http"
	"strconv"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/gorilla/mux"
)

// List returns all jobs.
func (s *Jobs) ListFunc(rw http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 0
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		offset = 0
	}

	jobsSlice, err := s.service.List(limit, offset)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	res := make([]jobs.JSONResponse, len(*jobsSlice))
	for i, job := range *jobsSlice {
		res[i] = job.ToJSONResponse()
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

// Details returns details regarding a job.
// It reads the job id for the wanted job from URL.
// Job service is responsible for validating the job id.
func (s *Jobs) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	job, err := s.service.Details(vars["jobId"])

	if err != nil {
		handleError(rw, r, err)
		return
	}

	res := job.ToJSONResponse()

	handleJsonResponse(rw, http.StatusOK, res)
}
