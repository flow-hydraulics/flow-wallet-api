package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Details returns details regarding a job.
// It reads the job id for the wanted job from URL.
// Job service is responsible for validating the job id.
func (s *Jobs) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := s.service.Details(vars["jobId"])

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
