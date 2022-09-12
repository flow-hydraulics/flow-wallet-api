package handlers

import (
	"net/http"
)

// InitMissingFungibleVaultsFunc starts job to init missing fungible token vaults.
func (s *Ops) InitMissingFungibleVaultsFunc(rw http.ResponseWriter, r *http.Request) {

	started, err := s.service.InitMissingFungibleTokenVaults()
	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, started)
}

// GetMissingFungibleVaultsFunc returns number of accounts with missing fungible token vaults.
func (s *Ops) GetMissingFungibleVaultsFunc(rw http.ResponseWriter, r *http.Request) {

	res, err := s.service.GetMissingFungibleTokenVaults()
	if err != nil {
		handleError(rw, r, err)
		return
	}

	handleJsonResponse(rw, http.StatusOK, res)
}
