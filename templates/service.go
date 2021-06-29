package templates

import (
	"fmt"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
)

type Service struct {
	store Store
	cfg   *config
}

func NewService(store Store) *Service {
	return &Service{store, parseConfig()}
}

func (s *Service) AddToken(t *Token) error {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(t.Address, s.cfg.ChainId)
	if err != nil {
		return err
	}

	t.Address = address

	if t.Name == "" {
		return fmt.Errorf(`not a valid name: "%s"`, t.Name)
	}

	return s.store.Insert(t)
}

func (s *Service) ListTokens(tType *TokenType) (*[]Token, error) {
	return s.store.List(tType)
}

func (s *Service) GetTokenById(id uint64) (*Token, error) {
	return s.store.GetById(id)
}

func (s *Service) GetTokenByName(name string) (*Token, error) {
	return s.store.GetByName(name)
}

func (s *Service) RemoveToken(id uint64) error {
	return s.store.Remove(id)
}
