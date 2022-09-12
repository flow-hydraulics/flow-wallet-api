package ops

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
}

// NewService initiates a new ops service.
func NewService(store Store) Service {
	return &ServiceImpl{store}
}
