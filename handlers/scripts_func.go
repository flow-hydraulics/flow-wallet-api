package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/scripts"
)

func (s *Scripts) ExecuteFunc(rw http.ResponseWriter, r *http.Request) {
	var err error

	if r.Body == nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		}
		handleError(rw, s.log, err)
		return
	}

	var b scripts.Script

	// Try to decode the request body into the struct.
	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid body"),
		}
		handleError(rw, s.log, err)
		return
	}

	res, err := s.service.Execute(r.Context(), b)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
