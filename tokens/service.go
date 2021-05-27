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

func (s *Service) CreateFtWithdrawalSync(
	ctx context.Context,
	token, sender, recipient, amount string,
) (*transactions.Transaction, error) {
	code, args, err := parseFtWithdrawal(s.cfg.ChainId, recipient, amount, token)
	if err != nil {
		return &transactions.EmptyTransaction, err
	}
	return s.ts.CreateSync(ctx, sender, code, args, transactions.Withdrawal)
}

func (s *Service) CreateFtWithdrawalAsync(
	token, sender, recipient, amount string,
) (*jobs.Job, error) {
	code, args, err := parseFtWithdrawal(s.cfg.ChainId, recipient, amount, token)
	if err != nil {
		return &jobs.Job{}, err
	}
	return s.ts.CreateAsync(sender, code, args, transactions.Withdrawal)
}
