package basic

import (
	"context"

	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
)

func (s *KeyManager) InitAdminProposalKeys(ctx context.Context) (uint16, error) {
	adminAddress := flow.HexToAddress(s.cfg.AdminAddress)

	adminAccount, err := s.fc.GetAccount(ctx, adminAddress)
	if err != nil {
		return 0, err
	}

	// TODO: we should not do this
	err = s.store.DeleteAllProposalKeys()
	if err != nil {
		return 0, err
	}

	var count uint16
	for _, k := range adminAccount.Keys {
		if !k.Revoked {
			err = s.store.InsertProposalKey(keys.ProposalKey{
				KeyIndex: k.Index,
			})
			if err != nil {
				return count, err
			}
			count += 1
		}
	}

	return count, nil
}
