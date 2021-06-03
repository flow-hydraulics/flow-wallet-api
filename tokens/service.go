package tokens

import (
	"context"

	"github.com/eqlabs/flow-wallet-service/jobs"
	"github.com/eqlabs/flow-wallet-service/transactions"
)

type Service struct {
	ts  *transactions.Service
	cfg Config
}

func NewService(ts *transactions.Service) *Service {
	cfg := ParseConfig()
	return &Service{ts, cfg}
}

func (s *Service) CreateFtWithdrawal(ctx context.Context, sync bool, tokenName, sender, recipient, amount string) (*jobs.Job, *transactions.Transaction, error) {
	raw, err := parseFtWithdrawal(tokenName, recipient, amount, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	return s.ts.Create(ctx, sync, sender, raw, transactions.FtWithdrawal)
}

func (s *Service) SetupFtForAccount(ctx context.Context, sync bool, tokenName, address string) (*jobs.Job, *transactions.Transaction, error) {
	raw, err := parseFtSetup(tokenName, s.cfg.ChainId)
	if err != nil {
		return nil, nil, err
	}
	return s.ts.Create(ctx, sync, address, raw, transactions.FtWithdrawal)
}
