package tokens

import (
	"context"
	"fmt"

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
	return s.ts.CreateSync(ctx, sender, code, args, transactions.FtWithdrawal)
}

func (s *Service) CreateFtWithdrawalAsync(tokenName, sender, recipient, amount string) (*jobs.Job, error) {
	code, args, err := parseFtWithdrawal(tokenName, recipient, amount, s.cfg.ChainId)
	if err != nil {
		return &jobs.Job{}, err
	}
	return s.ts.CreateAsync(sender, code, args, transactions.FtWithdrawal)
}

func (s *Service) SetupFtForAccountSync(ctx context.Context, tokenName, address string) (*transactions.Transaction, error) {
	code, args, err := parseFtSetup(tokenName, s.cfg.ChainId)
	if err != nil {
		return &transactions.EmptyTransaction, err
	}
	return s.ts.CreateSync(ctx, address, code, args, transactions.FtWithdrawal)
}

func (s *Service) SetupFtForAccountAsync(tokenName, address string) (*jobs.Job, error) {
	code, args, err := parseFtSetup(tokenName, s.cfg.ChainId)
	if err != nil {
		return &jobs.Job{}, err
	}
	fmt.Println(code, args)
	return s.ts.CreateAsync(address, code, args, transactions.FtWithdrawal)
}
