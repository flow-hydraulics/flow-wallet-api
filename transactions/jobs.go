package transactions

import (
	"context"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
)

const TransactionJobType = "transaction"

func (s *ServiceImpl) executeTransactionJob(ctx context.Context, j *jobs.Job) error {
	if j.Type != TransactionJobType {
		return jobs.ErrInvalidJobType
	}

	j.ShouldSendNotification = true

	tx, err := s.store.Transaction(j.TransactionID)
	if err != nil {
		return err
	}

	err = s.sendTransaction(ctx, &tx)
	if err != nil {
		return err
	}

	return nil
}
