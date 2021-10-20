package test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/chain_events"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

type Services interface {
	GetAccounts() *accounts.Service
	GetTokens() *tokens.Service
	GetTransactions() *transactions.Service

	GetKeyManager() keys.Manager
	GetListener() *chain_events.Listener
}

type svcs struct {
	accountService     *accounts.Service
	tokenService       *tokens.Service
	transactionService *transactions.Service

	keyManager keys.Manager
	listener   *chain_events.Listener
}

func GetServices(t *testing.T, cfg *configs.Config) Services {
	t.Helper()

	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dbClose := func() { gorm.Close(db) }
	t.Cleanup(dbClose)

	fc := NewFlowClient(t, cfg)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	templateStore := templates.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	wpStop := func() { wp.Stop() }
	t.Cleanup(wpStop)

	templateService := templates.NewService(cfg, templateStore)
	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService, templateService)
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, transactionService, templateService, accountService)

	store := chain_events.NewGormStore(db)
	getTypes := func() []string {
		// Get all enabled tokens
		tt, err := templateService.ListTokens(templates.NotSpecified)
		if err != nil {
			panic(err)
		}

		token_count := len(*tt)
		event_types := make([]string, token_count)

		// Listen for enabled tokens deposit events
		for i, token := range *tt {
			event_types[i] = templates.DepositEventTypeFromToken(token)
		}

		return event_types
	}

	le := log.New(os.Stdout, "[EVENT-POLLER] ", log.LstdFlags|log.Lshortfile)
	listener := chain_events.NewListener(
		le, fc, store, getTypes,
		cfg.ChainListenerMaxBlocks,
		1*time.Second,
		cfg.ChainListenerStartingHeight,
	)

	t.Cleanup(listener.Stop)

	// Register a handler for chain events
	chain_events.Event.Register(&tokens.ChainEventHandler{
		AccountService:  accountService,
		ChainListener:   listener,
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	listener.Start()

	ctx := context.Background()
	err = accountService.InitAdminAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// make sure all requested proposal keys are created
	keyCount, err := km.InitAdminProposalKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if keyCount != cfg.AdminProposalKeyCount {
		t.Fatal("incorrect number of admin proposal keys")
	}

	return &svcs{
		accountService:     accountService,
		tokenService:       tokenService,
		transactionService: transactionService,

		keyManager: km,
		listener:   listener,
	}
}

func (s *svcs) GetAccounts() *accounts.Service {
	return s.accountService
}

func (s *svcs) GetTokens() *tokens.Service {
	return s.tokenService
}

func (s *svcs) GetTransactions() *transactions.Service {
	return s.transactionService
}

func (s *svcs) GetKeyManager() keys.Manager {
	return s.keyManager
}

func (s *svcs) GetListener() *chain_events.Listener {
	return s.listener
}
