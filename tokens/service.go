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

func (s *Service) CreateFtWithdrawalSync(ctx context.Context, tokenName, sender, recipient, amount string) (*transactions.Transaction, error) {
	code, args, err := parseFtWithdrawal(tokenName, recipient, amount, s.cfg.ChainId)
	if err != nil {
		return &transactions.EmptyTransaction, err
	}
	_, t, err := s.ts.Create(ctx, true, sender, code, args, transactions.FtWithdrawal)
	return t, err
}

func (s *Service) CreateFtWithdrawalAsync(ctx context.Context, tokenName, sender, recipient, amount string) (*jobs.Job, error) {
	code, args, err := parseFtWithdrawal(tokenName, recipient, amount, s.cfg.ChainId)
	if err != nil {
		return &jobs.Job{}, err
	}
	j, _, err := s.ts.Create(ctx, false, sender, code, args, transactions.FtWithdrawal)
	return j, err
}

func (s *Service) SetupFtForAccountSync(ctx context.Context, tokenName, address string) (*transactions.Transaction, error) {
	code, args, err := parseFtSetup(tokenName, s.cfg.ChainId)
	if err != nil {
		return &transactions.EmptyTransaction, err
	}
	_, t, err := s.ts.Create(ctx, true, address, code, args, transactions.FtWithdrawal)
	return t, err
}

func (s *Service) SetupFtForAccountAsync(ctx context.Context, tokenName, address string) (*jobs.Job, error) {
	code, args, err := parseFtSetup(tokenName, s.cfg.ChainId)
	if err != nil {
		return &jobs.Job{}, err
	}
	j, _, err := s.ts.Create(ctx, false, address, code, args, transactions.FtWithdrawal)
	return j, err
}
