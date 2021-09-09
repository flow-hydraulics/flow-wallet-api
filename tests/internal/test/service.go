package test

import (
	"context"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/basic"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
)

type Services interface {
	GetAccounts() *accounts.Service
	GetTransactions() *transactions.Service
}

type svcs struct {
	accountService     *accounts.Service
	transactionService *transactions.Service
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
	txStore := transactions.NewGormStore(db)

	templateStore := templates.NewGormStore(db)
	templateService := templates.NewService(cfg, templateStore)

	km := basic.NewKeyManager(cfg, keyStore, fc)

	wp := jobs.NewWorkerPool(nil, jobStore, 100, 1)
	wpStop := func() { wp.Stop() }
	t.Cleanup(wpStop)

	txService := transactions.NewService(cfg, txStore, km, fc, wp)
	accountService := accounts.NewService(cfg, accountStore, km, fc, wp, txService, templateService)

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
		transactionService: txService,
	}
}

func (s *svcs) GetAccounts() *accounts.Service {
	return s.accountService
}

func (s *svcs) GetTransactions() *transactions.Service {
	return s.transactionService
}
