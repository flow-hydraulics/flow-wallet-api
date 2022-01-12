package tokens

import (
	"context"
	"encoding/json"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
)

const WithdrawalCreateJobType = "withdrawal_create"

type withdrawalCreateJobAttributes struct {
	Sender  string
	Request WithdrawalRequest
}

func (s *ServiceImpl) executeCreateWithdrawalJob(ctx context.Context, j *jobs.Job) error {
	if j.Type != WithdrawalCreateJobType {
		return jobs.ErrInvalidJobType
	}

	j.ShouldSendNotification = true

	attrs := withdrawalCreateJobAttributes{}
	if err := json.Unmarshal(j.Attributes, &attrs); err != nil {
		return err
	}

	transaction, err := s.createWithdrawal(ctx, attrs.Sender, attrs.Request)
	if err != nil {
		return err
	}

	j.TransactionID = transaction.TransactionId
	j.Result = transaction.TransactionId

	return nil
}
