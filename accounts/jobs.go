package accounts

import (
	"context"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
)

const AccountCreateJobType = "account_create"

func (s *Service) executeAccountCreateJob(ctx context.Context, j *jobs.Job) error {
	if j.Type != AccountCreateJobType {
		return jobs.ErrInvalidJobType
	}

	j.ShouldSendNotification = true

	a, txID, err := s.createAccount(ctx)
	if err != nil {
		return err
	}

	j.TransactionID = txID
	j.Result = a.Address

	return nil
}
