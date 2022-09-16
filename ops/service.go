package ops

import "github.com/flow-hydraulics/flow-wallet-api/templates"

// Service lists all functionality provided by ops service
type Service interface {
	// Retroactive fungible token vault initialization
	GetMissingFungibleTokenVaults() ([]TokenCount, error)
	InitMissingFungibleTokenVaults() (bool, error)
}

type TokenCount struct {
	TokenName string `json:"token"`
	Count     uint   `json:"count"`
}

// ServiceImpl implements the ops Service
type ServiceImpl struct {
	store Store
	temps templates.Service
}

// NewService initiates a new ops service.
func NewService(
	store Store,
	temps templates.Service,
) Service {
	return &ServiceImpl{store, temps}
}
