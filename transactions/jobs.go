package transactions

import (
	"context"
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
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

	latestSequenceNum := acc.Keys[flowTx.ProposalKey.KeyIndex].SequenceNumber
	if flowTx.ProposalKey.SequenceNumber != latestSequenceNum {
		log.Warnf("Sequence number mismatch, updating transaction: jobID=%s", j.ID)
		newTx, err := s.rebuildTransaction(ctx, &tx, flowTx)
		if err != nil {
			return err
		}
		if err := s.store.InsertTransaction(newTx); err != nil {
			return fmt.Errorf("error while creating new transaction: %w", err)
		}
		j.TransactionID = newTx.TransactionId
		if err := s.jobStore.UpdateJob(j); err != nil {
			return fmt.Errorf("error while updating job: %w", err)
		}
		tx = *newTx
	}

	err = s.sendTransaction(ctx, &tx)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceImpl) rebuildTransaction(ctx context.Context, tx *Transaction, flowTx *flow.Transaction) (*Transaction, error) {
	args := make([]Argument, len(flowTx.Arguments))
	for i := range flowTx.Arguments {
		arg, err := flowTx.Argument(i)
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}
	newFlowTx, err := s.buildFlowTransaction(ctx, flowTx.ProposalKey.Address.String(), string(flowTx.Script), args)
	if err != nil {
		return nil, err
	}
	newTx := &Transaction{
		TransactionId:   newFlowTx.ID().Hex(),
		FlowTransaction: newFlowTx.Encode(),
		ProposerAddress: tx.ProposerAddress,
		TransactionType: tx.TransactionType,
	}
	return newTx, nil
}
