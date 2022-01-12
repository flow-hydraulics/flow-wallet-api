package accounts

import (
	"context"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"

	flow_templates "github.com/onflow/flow-go-sdk/templates"
)

// AddContract is used only in tests
func AddContract(
	ctx context.Context,
	fc flow_helpers.FlowClient,
	km keys.Manager,
	accountAddress string,
	contract flow_templates.Contract,
	transactionTimeout time.Duration) error {

	// Get admin account authorizer
	payer, err := km.AdminAuthorizer(ctx)
	if err != nil {
		return err
	}

	// Get user account authorizer
	proposer, err := km.UserAuthorizer(ctx, flow.HexToAddress(accountAddress))
	if err != nil {
		return err
	}

	// Get latest blocks id as reference id
	referenceBlockID, err := flow_helpers.LatestBlockId(ctx, fc)
	if err != nil {
		return err
	}

	flowTx := flow.NewTransaction().
		SetReferenceBlockID(*referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address).
		SetGasLimit(maxGasLimit).
		SetScript([]byte(template_strings.AddAccountContractWithAdmin)).
		AddAuthorizer(payer.Address)

	if err := flowTx.AddArgument(cadence.String(contract.Name)); err != nil {
		return err
	}

	if err := flowTx.AddArgument(cadence.String(contract.SourceHex())); err != nil {
		return err
	}

	// Proposer signs the payload
	if proposer.Address.Hex() != payer.Address.Hex() {
		if err := flowTx.SignPayload(proposer.Address, proposer.Key.Index, proposer.Signer); err != nil {
			return err
		}
	}

	// Payer signs the envelope
	if err := flowTx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer); err != nil {
		return err
	}

	_, err = flow_helpers.SendAndWait(ctx, fc, *flowTx, transactionTimeout)
	if err != nil {
		return err
	}

	return nil
}
