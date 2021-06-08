package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/gorilla/mux"
)

func (s *FungibleTokens) CreateWithdrawalFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sender := vars["address"]
	tN := vars["tokenName"]

	var b FTWithdrawalRequest

	if r.Body == nil || r.Body == http.NoBody {
		err := &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("empty body")}
		handleError(rw, s.log, err)
		return
	}

	// Try to decode the request body.
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		err = &errors.RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("invalid body")}
		handleError(rw, s.log, err)
		return
	}

	b.Name = tN

	// Decide whether to serve sync or async, default async
	sync := r.Header.Get(SYNC_HEADER) != ""
	job, tx, err := s.service.CreateFtWithdrawal(r.Context(), sync, b.Token, sender, b.Recipient, b.Amount)
	var res interface{}
	if sync {
		res = tx
	} else {
		res = job
	}

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, res)
}

func (s *FungibleTokens) DetailsFunc(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)

	a := vars["address"]
	tN := vars["tokenName"]
	tA := r.Form.Get("tokenAddress")

	t := templates.NewToken(tN, tA)

	res, err := s.service.Details(r.Context(), t, a)

	if err != nil {
		handleError(rw, s.log, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
