package ops

import (
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

// Service lists all functionality provided by ops service
type Service interface {
	// Retroactive fungible token vault initialization
	GetMissingFungibleTokenVaults() ([]TokenCount, error)
	InitMissingFungibleTokenVaults() (string, error)
	GetWorkerPool() OpsWorkerPoolService
}

// ServiceImpl implements the ops Service
type ServiceImpl struct {
	cfg    *configs.Config
	store  Store
	temps  templates.Service
	txs    transactions.Service
	tokens tokens.Service
	wp     OpsWorkerPoolService

	initFungibleJobRunning bool
}

// NewService initiates a new ops service.
func NewService(
	cfg *configs.Config,
	store Store,
	temps templates.Service,
	txs transactions.Service,
	tokens tokens.Service,
) Service {

	wp := NewWorkerPool(
		cfg.OpsWorkerCount,
		cfg.OpsWorkerQueueCapacity,
	)
	wp.Start()

	return &ServiceImpl{cfg, store, temps, txs, tokens, wp, false}
}

func (s *ServiceImpl) GetWorkerPool() OpsWorkerPoolService {
	return s.wp
}
