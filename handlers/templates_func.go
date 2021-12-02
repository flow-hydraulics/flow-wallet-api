package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/gorilla/mux"
)

func (s *Templates) AddTokenFunc(rw http.ResponseWriter, r *http.Request) {
	// TODO (latenssi): separate request, response and db structs
	var newToken templates.Token

	// Check body is not empty
	if err := checkNonEmptyBody(r); err != nil {
		handleError(rw, r, err)
		return
	}

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&newToken); err != nil {
		handleError(rw, r, InvalidBodyError)
		return
	}

	// Help marshalling the ID field if user for some reason tried to send a create request with an ID
	newToken.ID = 0

	// Add the token to database
	if err := s.service.AddToken(&newToken); err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusCreated, newToken)
}

func (s *Templates) MakeListTokensFunc(tType templates.TokenType) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		tokens, err := s.service.ListTokens(tType)
		if err != nil {
			handleError(rw, r, err)
			return
		}
		handleJsonResponse(rw, http.StatusOK, tokens)
	}
}

func (s *Templates) GetTokenFunc(rw http.ResponseWriter, r *http.Request) {
	var token *templates.Token
	var err error

	vars := mux.Vars(r)
	idOrName := vars["id_or_name"]

	id, err := strconv.ParseUint(idOrName, 10, 64)

	if err == nil {
		token, err = s.service.GetTokenById(id)
		if err != nil {
			handleError(rw, r, err)
			return
		}
	} else {
		token, err = s.service.GetTokenByName(idOrName)
		if err != nil {
			handleError(rw, r, err)
			return
		}
	}

	handleJsonResponse(rw, http.StatusOK, token)
}

func (s *Templates) RemoveTokenFunc(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		handleError(rw, r, err)
		return
	}

	if err := s.service.RemoveToken(id); err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, id)
}
