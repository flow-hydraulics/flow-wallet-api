package handlers

import (
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/ops"
)

// Ops is a HTTP server for admin operations.
type Ops struct {
	service ops.Service
}

// NewJobs initiates a new jobs server.
func NewOps(service ops.Service) *Ops {
	return &Ops{service}
}

func (s *Ops) InitMissingFungibleVaults() http.Handler {
	return http.HandlerFunc(s.InitMissingFungibleVaultsFunc)
}

func (s *Ops) GetMissingFungibleVaults() http.Handler {
	return http.HandlerFunc(s.GetMissingFungibleVaultsFunc)
}
