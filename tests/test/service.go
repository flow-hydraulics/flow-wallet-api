package test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
	upstreamgorm "gorm.io/gorm"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/chain_events"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/ops"
	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

type Services interface {
	GetAccounts() accounts.Service
	GetJobs() jobs.Service
	GetTemplates() templates.Service
	GetTokens() tokens.Service
	GetTransactions() transactions.Service
	GetSystem() system.Service
	GetOps() ops.Service

	GetKeyManager() keys.Manager
	GetListener() chain_events.Listener
	GetFlowClient() flow_helpers.FlowClient
}

type svcs struct {
	accountService     accounts.Service
	jobService         jobs.Service
	templateService    templates.Service
	tokenService       tokens.Service
	transactionService transactions.Service
	systemService      system.Service
	opsService         ops.Service

	keyManager keys.Manager
	listener   chain_events.Listener
	flowClient flow_helpers.FlowClient
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
			if err := m.DropTable(table); err != nil {
				t.Logf("error while cleaning test database: %s", err)
			}
		}
	}
	t.Cleanup(dbClose)
	t.Cleanup(dbClean)

	return db
}

func GetServices(t *testing.T, cfg *configs.Config) Services {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"), // Ignore OpenCensus
			goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),           // Ignore goroutine leak from AWS KMS
			goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),              // Ignore goroutine leak from AWS KMS
			goleak.IgnoreTopFunction("database/sql.(*DB).connectionOpener"),
			goleak.IgnoreTopFunction("google.golang.org/grpc.(*ccBalancerWrapper).watcher"),
			goleak.IgnoreTopFunction("google.golang.org/grpc/internal/transport.(*controlBuffer).get"),
			goleak.IgnoreTopFunction("github.com/flow-hydraulics/flow-wallet-api/jobs.(*WorkerPoolImpl).startWorkers.func1"),
			goleak.IgnoreTopFunction("github.com/flow-hydraulics/flow-wallet-api/ops.(*workerPoolImpl).Start.func1"),
			goleak.IgnoreTopFunction("github.com/flow-hydraulics/flow-wallet-api/jobs.(*WorkerPoolImpl).startDBJobScheduler.func1"),
			goleak.IgnoreTopFunction("github.com/flow-hydraulics/flow-wallet-api/chain_events.(*ListenerImpl).Start.func1"),
		)
	})

	db := GetDatabase(t, cfg)
	fc := NewFlowClient(t, cfg)

	systemService := system.NewService(system.NewGormStore(db))

	wp := jobs.NewWorkerPool(
		jobs.NewGormStore(db),
		cfg.WorkerQueueCapacity,
		cfg.WorkerCount,
		jobs.WithMaxJobErrorCount(0),
		jobs.WithDbJobPollInterval(time.Second),
		jobs.WithAcceptedGracePeriod(1000),
		jobs.WithReSchedulableGracePeriod(1000),
		jobs.WithSystemService(systemService),
	)

	km := basic.NewKeyManager(cfg, keys.NewGormStore(db), fc)

	templateService, err := templates.NewService(cfg, templates.NewGormStore(db))
	if err != nil {
		t.Fatal(err)
	}
	transactionService := transactions.NewService(cfg, transactions.NewGormStore(db), km, fc, wp)
	accountService := accounts.NewService(cfg, accounts.NewGormStore(db), km, fc, wp, transactionService, templateService)
	jobService := jobs.NewService(jobs.NewGormStore(db))
	tokenService := tokens.NewService(cfg, tokens.NewGormStore(db), km, fc, wp, transactionService, templateService, accountService)
	opsService := ops.NewService(cfg, ops.NewGormStore(db), templateService, transactionService, tokenService)

	getTypes := func() ([]string, error) {
		// Get all enabled tokens
		tt, err := templateService.ListTokens(templates.NotSpecified)
		if err != nil {
			return nil, err
		}

		token_count := len(tt)
		event_types := make([]string, token_count)

		// Listen for enabled tokens deposit events
		for i, token := range tt {
			event_types[i] = templates.DepositEventTypeFromToken(token)
		}

		return event_types, nil
	}

	listener := chain_events.NewListener(
		fc, chain_events.NewGormStore(GetDatabase(t, cfg)), getTypes,
		cfg.ChainListenerMaxBlocks,
		1*time.Second,
		cfg.ChainListenerStartingHeight,
	)

	// Register a handler for chain events
	chain_events.ChainEvent.Register(&tokens.ChainEventHandler{
		AccountService:  accountService,
		ChainListener:   listener,
		TemplateService: templateService,
		TokenService:    tokenService,
	})

	err = accountService.InitAdminAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// make sure all requested proposal keys are created
	if err := km.CheckAdminProposalKeyCount(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		wp.Stop(false)
		opsService.GetWorkerPool().Stop()
		listener.Stop()
	})

	wp.Start()
	listener.Start()

	return &svcs{
		accountService:     accountService,
		jobService:         jobService,
		templateService:    templateService,
		tokenService:       tokenService,
		transactionService: transactionService,
		systemService:      systemService,
		opsService:         opsService,

		keyManager: km,
		listener:   listener,
		flowClient: fc,
	}
}

func (s *svcs) GetAccounts() accounts.Service {
	return s.accountService
}

func (s *svcs) GetJobs() jobs.Service {
	return s.jobService
}

func (s *svcs) GetTemplates() templates.Service {
	return s.templateService
}

func (s *svcs) GetTokens() tokens.Service {
	return s.tokenService
}

func (s *svcs) GetTransactions() transactions.Service {
	return s.transactionService
}

func (s *svcs) GetOps() ops.Service {
	return s.opsService
}

func (s *svcs) GetKeyManager() keys.Manager {
	return s.keyManager
}

func (s *svcs) GetListener() chain_events.Listener {
	return s.listener
}

func (s *svcs) GetSystem() system.Service {
	return s.systemService
}

func (s *svcs) GetFlowClient() flow_helpers.FlowClient {
	return s.flowClient
}
