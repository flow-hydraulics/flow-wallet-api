package templates

import (
	"fmt"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
)

type Service struct {
	db  Store
	cfg *config
}

func NewService(db Store) *Service {
	return &Service{db, parseConfig()}
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

	return s.db.Insert(t)
}

func (s *Service) ListTokens() (*[]Token, error) {
	return s.db.List()
}

func (s *Service) GetToken(id uint64) (*Token, error) {
	return s.db.Get(id)
}

func (s *Service) RemoveToken(id uint64) error {
	return s.db.Remove(id)
}
