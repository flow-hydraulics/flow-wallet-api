package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/gorilla/mux"
)

func (s *Transactions) ListFunc(rw http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 0
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		offset = 0
	}

	vars := mux.Vars(r)

	res, err := s.service.List(vars["address"], limit, offset)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Transactions) CreateFunc(rw http.ResponseWriter, r *http.Request) {
	var err error

	if r.Body == nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		}
		handleError(rw, s.log, err)
		return
	}

	vars := mux.Vars(r)

	var t transactions.Transaction

	// Try to decode the request body into the struct.
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid body"),
		}
		handleError(rw, s.log, err)
		return
	}

	// Decide whether to serve sync or async, default async
	var res interface{}
	if us := r.Header.Get(SYNC_HEADER); us != "" {
		res, err = s.service.CreateSync(r.Context(), t.Code, t.Arguments, vars["address"])
	} else {
		res, err = s.service.CreateAsync(t.Code, t.Arguments, vars["address"])
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *Transactions) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := s.service.Details(vars["address"], vars["transactionId"])

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}

func (s *Transactions) ExecuteScriptFunc(rw http.ResponseWriter, r *http.Request) {
	var err error

	if r.Body == nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("empty body"),
		}
		handleError(rw, s.log, err)
		return
	}

	var t transactions.Script

	// Try to decode the request body into the struct.
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("invalid body"),
		}
		handleError(rw, s.log, err)
		return
	}

	res, err := s.service.ExecuteScript(r.Context(), t.Code, t.Arguments)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
