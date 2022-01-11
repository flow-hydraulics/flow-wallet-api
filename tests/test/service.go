package test

import (
	"context"
	"testing"
	"time"

	upstreamgorm "gorm.io/gorm"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/chain_events"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

type Services interface {
	GetAccounts() *accounts.Service
	GetJobs() *jobs.Service
	GetTokens() *tokens.Service
	GetTransactions() *transactions.Service
	GetSystem() *system.Service

	GetKeyManager() keys.Manager
	GetListener() *chain_events.Listener
}

type svcs struct {
	accountService     *accounts.Service
	jobService         *jobs.Service
	tokenService       *tokens.Service
	transactionService *transactions.Service
	systemService      *system.Service

	keyManager keys.Manager
	listener   *chain_events.Listener
}

func GetDatabase(t *testing.T, cfg *configs.Config) *upstreamgorm.DB {
	db, err := gorm.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	dbClose := func() { gorm.Close(db) }
	dbClean := func() {
		m := db.Migrator()
		tables, err := m.GetTables()
		if err != nil {
			t.Logf("error while cleaning test database: %s", err)
		}
		for _, table := range tables {
			m.DropTable(table)
		}
	}
	t.Cleanup(dbClose)
	t.Cleanup(dbClean)

	return db
}

func GetServices(t *testing.T, cfg *configs.Config) Services {

	db := GetDatabase(t, cfg)
	fc := NewFlowClient(t, cfg)

	jobStore := jobs.NewGormStore(db)
	accountStore := accounts.NewGormStore(db)
	keyStore := keys.NewGormStore(db)
	templateStore := templates.NewGormStore(db)
	tokenStore := tokens.NewGormStore(db)
	transactionStore := transactions.NewGormStore(db)
	systemStore := system.NewGormStore(db)
	listenerStore := chain_events.NewGormStore(GetDatabase(t, cfg))

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(
		jobStore,
		cfg.WorkerQueueCapacity,
		cfg.WorkerCount,
		jobs.WithMaxJobErrorCount(0),
		jobs.WithDbJobPollInterval(time.Second),
		jobs.WithAcceptedGracePeriod(0),
		jobs.WithReSchedulableGracePeriod(0),
	)

	templateService := templates.NewService(cfg, templateStore)
	transactionService := transactions.NewService(cfg, transactionStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, transactionService)
	jobService := jobs.NewService(jobStore)
	tokenService := tokens.NewService(cfg, tokenStore, km, fc, wp, transactionService, templateService, accountService)
	systemService := system.NewService(systemStore)

	getTypes := func() ([]string, error) {
		// Get all enabled tokens
		tt, err := templateService.ListTokens(templates.NotSpecified)
		if err != nil {
			return nil, err
		}

		token_count := len(*tt)
		event_types := make([]string, token_count)

		// Listen for enabled tokens deposit events
		for i, token := range *tt {
			event_types[i] = templates.DepositEventTypeFromToken(token)
		}

		return event_types, nil
	}

	listener := chain_events.NewListener(
		fc, listenerStore, getTypes,
		cfg.ChainListenerMaxBlocks,
		1*time.Second,
		cfg.ChainListenerStartingHeight,
	)

	// Register a handler for chain events
	chain_events.Event.Register(&tokens.ChainEventHandler{
		AccountService:  accountService,
		ChainListener:   listener,
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	err := accountService.InitAdminAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// make sure all requested proposal keys are created
	if err := km.CheckAdminProposalKeyCount(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		wp.Stop(false)
		listener.Stop()
	})

	wp.Start()
	listener.Start()

	return &svcs{
		accountService:     accountService,
		jobService:         jobService,
		tokenService:       tokenService,
		transactionService: transactionService,
		systemService:      systemService,

		keyManager: km,
		listener:   listener,
	}
}

func (s *svcs) GetAccounts() *accounts.Service {
	return s.accountService
}

func (s *svcs) GetJobs() *jobs.Service {
	return s.jobService
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

func (s *svcs) GetSystem() *system.Service {
	return s.systemService
}
