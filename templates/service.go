package templates

import (
	"fmt"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	AddToken(t *Token) error
	ListTokens(tType TokenType) ([]BasicToken, error)
	ListTokensFull(tType TokenType) ([]Token, error)
	GetTokenById(id uint64) (*Token, error)
	GetTokenByName(name string) (*Token, error)
	RemoveToken(id uint64) error
	TokenFromEvent(e flow.Event) (*Token, error)
}

type ServiceImpl struct {
	store Store
	cfg   *configs.Config
}

func parseEnabledTokens(envEnabledTokens []string) map[string]Token {
	var enabledTokens = make(map[string]Token, len(envEnabledTokens))
	for _, s := range envEnabledTokens {
		ss := strings.Split(s, ":")
		token := Token{Name: ss[0], Address: ss[1]}
		if len(ss) == 3 {
			// Deprecated
			if token.Name != "FlowToken" && token.Name != "FUSD" {
				log.Warnf("ENABLED_TOKENS.%s is using deprecated config format: %s", ss[0], s)
			}
			token.NameLowerCase = ss[2]
			token.ReceiverPublicPath = fmt.Sprintf("/public/%sReceiver", token.NameLowerCase)
			token.BalancePublicPath = fmt.Sprintf("/public/%sBalance", token.NameLowerCase)
			token.VaultStoragePath = fmt.Sprintf("/storage/%sVault", token.NameLowerCase)
		} else if len(ss) == 5 {
			token.ReceiverPublicPath = ss[2]
			token.BalancePublicPath = ss[3]
			token.VaultStoragePath = ss[4]
		}
		// Use all lowercase as the key so we can do case-insensitive matching in URLs
		key := strings.ToLower(ss[0])
		enabledTokens[key] = token
	}
	return enabledTokens
}

func NewService(cfg *configs.Config, store Store) (Service, error) {
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
				return nil, err
			}
		}

		// Copy the value so we get an individual pointer, this is important
		token := t
		token.Type = FT // We only allow fungible tokens through env variables config

		var err error

		token.Setup, err = FungibleSetupCode(cfg.ChainID, &token)
		if err != nil {
			return nil, err
		}

		token.Transfer, err = FungibleTransferCode(cfg.ChainID, &token)
		if err != nil {
			return nil, err
		}

		token.Balance, err = FungibleBalanceCode(cfg.ChainID, &token)
		if err != nil {
			return nil, err
		}

		// Write to temp storage (memory), instead of database
		store.InsertTemp(&token)
	}

	return &ServiceImpl{store, cfg}, nil
}

func (s *ServiceImpl) AddToken(t *Token) error {
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
	t.Setup, err = TokenCode(s.cfg.ChainID, t, t.Setup)
	if err != nil {
		return err
	}

	t.Transfer, err = TokenCode(s.cfg.ChainID, t, t.Transfer)
	if err != nil {
		return err
	}

	t.Balance, err = TokenCode(s.cfg.ChainID, t, t.Balance)
	if err != nil {
		return err
	}

	return s.store.Insert(t)
}

func (s *ServiceImpl) ListTokens(tType TokenType) ([]BasicToken, error) {
	return s.store.List(tType)
}

func (s *ServiceImpl) ListTokensFull(tType TokenType) ([]Token, error) {
	return s.store.ListFull(tType)
}

func (s *ServiceImpl) GetTokenById(id uint64) (*Token, error) {
	return s.store.GetById(id)
}

func (s *ServiceImpl) GetTokenByName(name string) (*Token, error) {
	return s.store.GetByName(name)
}

func (s *ServiceImpl) RemoveToken(id uint64) error {
	return s.store.Remove(id)
}

func (s *ServiceImpl) TokenFromEvent(e flow.Event) (*Token, error) {
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
