package handlers

import (
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/ops"
)

// Ops is a HTTP server for admin (system) operations.
type Ops struct {
	service ops.Service
}

// NewOps initiates a new ops server.
func NewOps(service ops.Service) *Ops {
	return &Ops{service}
}

// InitMissingFungibleVaults starts the job to initialize missing fungible token vaults.
func (s *Ops) InitMissingFungibleVaults() http.Handler {
	return http.HandlerFunc(s.InitMissingFungibleVaultsFunc)
}

// GetMissingFungibleVaults returns number of accounts that are missing a configured fungible token vault.
func (s *Ops) GetMissingFungibleVaults() http.Handler {
	return http.HandlerFunc(s.GetMissingFungibleVaultsFunc)
}
