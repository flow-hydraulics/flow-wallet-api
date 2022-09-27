package ops

import (
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

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
	cfg   *configs.Config
	store Store
	temps templates.Service
	txs   transactions.Service
}

// NewService initiates a new ops service.
func NewService(
	cfg *configs.Config,
	store Store,
	temps templates.Service,
	txs transactions.Service,
) Service {
	return &ServiceImpl{cfg, store, temps, txs}
}
