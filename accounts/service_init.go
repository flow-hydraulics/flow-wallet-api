package accounts

import (
	"context"
	"errors"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

func (s *ServiceImpl) InitAdminAccount(ctx context.Context) error {
	log.Debug("Initializing admin account")

	a, err := s.store.Account(s.cfg.AdminAddress)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return err
		}
		// Admin account not in database
		a = Account{Address: s.cfg.AdminAddress}
		err := s.store.InsertAccount(&a)
		if err != nil {
			return err
		}
		AccountAdded.Trigger(AccountAddedPayload{
			Address: flow.HexToAddress(s.cfg.AdminAddress),
		})
	}

	log.WithFields(log.Fields{
		"wantedCount": s.cfg.AdminProposalKeyCount,
	}).Debug("Checking admin account proposal keys")

	if err := s.km.CheckAdminProposalKeyCount(ctx); err != nil {
		if errors.Is(err, keys.ErrAdminProposalKeyCountMismatch) {
			log.WithFields(log.Fields{
				"err": err,
			}).Info("Admin proposal key count mismatch")

			keyCount, err := s.km.InitAdminProposalKeys(ctx)
			if err != nil {
				return err
			}

			if keyCount < s.cfg.AdminProposalKeyCount {
				if err := s.addAdminProposalKeys(ctx, s.cfg.AdminProposalKeyCount-keyCount); err != nil {
					return err
				}

				if _, err := s.km.InitAdminProposalKeys(ctx); err != nil {
					return err
				}

				log.Info("New admin account proposal keys created successfully")
			}
		} else {
			return err
		}
	}

	return nil
}

func (s *ServiceImpl) addAdminProposalKeys(ctx context.Context, count uint16) error {

	log.
		WithFields(log.Fields{"count": count}).
		Info("Adding admin account proposal keys")

	payer, err := s.km.AdminAuthorizer(ctx)
	if err != nil {
		return err
	}

	referenceBlockID, err := flow_helpers.LatestBlockId(ctx, s.fc)
	if err != nil {
		return err
	}

	code := template_strings.AddProposalKeyTransaction

	flowTx := flow.NewTransaction()
	flowTx.
		SetReferenceBlockID(*referenceBlockID).
		SetProposalKey(payer.Address, payer.Key.Index, payer.Key.SequenceNumber).
		SetPayer(payer.Address).
		SetGasLimit(maxGasLimit).
		SetScript([]byte(code))

	if err := flowTx.AddArgument(cadence.NewInt(s.cfg.AdminKeyIndex)); err != nil {
		return err
	}

	if err := flowTx.AddArgument(cadence.NewUInt16(count)); err != nil {
		return err
	}

	flowTx.AddAuthorizer(payer.Address)

	if err := flowTx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer); err != nil {
		return err
	}

	// Send and wait for the transaction to be sealed
	if _, err := flow_helpers.SendAndWait(ctx, s.fc, *flowTx, s.cfg.TransactionTimeout); err != nil {
		return err
	}

	return nil
}
