package transactions

import (
	"context"
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/onflow/flow-go-sdk"
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

	flowTx, err := flow.DecodeTransaction(tx.FlowTransaction)
	if err != nil {
		return err
	}

	acc, err := s.fc.GetAccount(ctx, flowTx.ProposalKey.Address)
	if err != nil {
		return err
	}
	if len(acc.Keys) <= flowTx.ProposalKey.KeyIndex {
		return fmt.Errorf("proposal key index mismatch")
	}

	// TODO rebuild tx
	latestSequenceNum := acc.Keys[flowTx.ProposalKey.KeyIndex].SequenceNumber
	if flowTx.ProposalKey.SequenceNumber != latestSequenceNum {
		tx.FlowTransaction = flowTx.
			SetProposalKey(flowTx.ProposalKey.Address, flowTx.ProposalKey.KeyIndex, latestSequenceNum).
			Encode()
		if err := s.store.UpdateTransaction(&tx); err != nil {
			return fmt.Errorf("error while updating transaction in db: %w", err)
		}
	}

	err = s.sendTransaction(ctx, &tx)
	if err != nil {
		return err
	}

	return nil
}
