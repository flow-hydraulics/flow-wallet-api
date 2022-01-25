package accounts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

const AccountCreateJobType = "account_create"

func (s *ServiceImpl) executeAccountCreateJob(ctx context.Context, j *jobs.Job) error {
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

const SyncAccountKeyCountJobType = "sync_account_key_count"

type syncAccountKeyCountJobAttributes struct {
	Address flow.Address `json:"address"`
	NumKeys int          `json:"numkeys"`
}

func (s *ServiceImpl) executeSyncAccountKeyCountJob(ctx context.Context, j *jobs.Job) error {
	entry := log.WithFields(log.Fields{"job": j, "function": "executeSyncAccountKeyCountJob"})
	if j.Type != SyncAccountKeyCountJobType {
		return jobs.ErrInvalidJobType
	}

	j.ShouldSendNotification = true

	var attrs syncAccountKeyCountJobAttributes
	err := json.Unmarshal(j.Attributes, &attrs)
	if err != nil {
		return err
	}

	entry.WithFields(log.Fields{"attrs": j.Attributes}).Trace("Unmarshaled attributes")

	numKeys, txID, err := s.syncAccountKeyCount(ctx, attrs.Address, attrs.NumKeys)
	entry.WithFields(log.Fields{"numKeys": numKeys, "txId": txID, "err": err}).Trace("s.syncAccountKeyCount complete")
	if err != nil {
		return err
	}

	j.TransactionID = txID
	j.Result = fmt.Sprintf("%s:%d", attrs.Address, numKeys)

	return nil
}
