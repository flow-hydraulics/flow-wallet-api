package templates

import (
	"fmt"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	store Store
	cfg   *configs.Config
}

func parseEnabledTokens(envEnabledTokens []string) map[string]Token {
	var enabledTokens = make(map[string]Token, len(envEnabledTokens))
	for _, s := range envEnabledTokens {
		ss := strings.Split(s, ":")
		token := Token{Name: ss[0], Address: ss[1]}
		if len(ss) > 2 {
			token.NameLowerCase = ss[2]
		}
		// Use all lowercase as the key so we can do case insenstive matchig in URLs
		key := strings.ToLower(ss[0])
		enabledTokens[key] = token
	}
	return enabledTokens
}

func NewService(cfg *configs.Config, store Store) *Service {
	// TODO(latenssi): safeguard against nil config?

	// Add all enabled tokens from config as fungible tokens
	for _, t := range parseEnabledTokens(cfg.EnabledTokens) {
		if _, err := store.GetByName(t.Name); err == nil {
			// Token already in database
			log.
				WithFields(log.Fields{"error": err, "tokenName": t.Name}).
				Warn("Skipping configuration from environment variables as it already exists in database. Consider removing it from database or from environment variables.")
			continue
		} else {
			if !strings.Contains(err.Error(), "record not found") {
				// We got an error that is not "record not found"
				panic(err)
			}
		}

		// Copy the value so we get an individual pointer, this is important
		token := t
		token.Type = FT // We only allow fungible tokens through env variables config
		token.Setup = FungibleSetupCode(cfg.ChainID, &token)
		token.Transfer = FungibleTransferCode(cfg.ChainID, &token)
		token.Balance = FungibleBalanceCode(cfg.ChainID, &token)

		// Write to temp storage (memory), instead of database
		store.InsertTemp(&token)
	}

	return &Service{store, cfg}
}

func (s *Service) AddToken(t *Token) error {
	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(t.Address, s.cfg.ChainID)
	if err != nil {
		return err
	}

	t.Address = address

	if t.Name == "" {
		return fmt.Errorf(`not a valid name: "%s"`, t.Name)
	}

	// Received code templates may have values that need replacing
	t.Setup = TokenCode(s.cfg.ChainID, t, t.Setup)
	t.Transfer = TokenCode(s.cfg.ChainID, t, t.Transfer)
	t.Balance = TokenCode(s.cfg.ChainID, t, t.Balance)

	return s.store.Insert(t)
}

func (s *Service) ListTokens(tType TokenType) (*[]BasicToken, error) {
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

func (s *Service) TokenFromEvent(e flow.Event) (*Token, error) {
	// Example event:
	// A.0ae53cb6e3f42a79.FlowToken.TokensDeposited
	ss := strings.Split(e.Type, ".")

	// Token address from event
	eAddress, err := flow_helpers.ValidateAddress(ss[1], s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	token, err := s.GetTokenByName(ss[2])
	if err != nil {
		return nil, err
	}

	// Token address from database
	tAddress, err := flow_helpers.ValidateAddress(token.Address, s.cfg.ChainID)
	if err != nil {
		return nil, err
	}

	// Check if addresses match
	if eAddress != tAddress {
		return nil, fmt.Errorf("addresses do not match for %s, from event: %s, from database: %s", token.Name, eAddress, tAddress)
	}

	return token, nil
}
