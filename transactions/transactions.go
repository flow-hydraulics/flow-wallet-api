package transactions

import (
	"context"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

// Transaction is the database model for all transactions.
type Transaction struct {
	TransactionId   string                  `json:"transactionId" gorm:"column:transaction_id;primaryKey"`
	TransactionType Type                    `json:"transactionType" gorm:"column:transaction_type;index"`
	ProposerAddress string                  `json:"-" gorm:"column:proposer_address;index"`
	CreatedAt       time.Time               `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time               `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt          `json:"-" gorm:"column:deleted_at;index"`
	Result          *flow.TransactionResult `json:"-" gorm:"-"`
	Actual          *flow.Transaction       `json:"-" gorm:"-"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// TODO(latenssi): separate HTTP interface model for transactions

func New(
	transaction *Transaction,
	referenceBlockID flow.Identifier,
	builder *templates.TransactionBuilder,
	tType Type,
	proposer, payer keys.Authorizer,
	authorizers []keys.Authorizer) error {

	// TODO: Gas limit?
	builder.Tx.
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address)

	// Add authorizers
	for _, a := range authorizers {
		builder.Tx.AddAuthorizer(a.Address)
	}

	// Authorizers sign the payload
	for _, a := range authorizers {
		// If account is also the payer, it must only sign the envelope,
		// proposer signing is handled outside this loop as well
		if a.Equals(proposer) || a.Equals(payer) {
			continue
		}

		if err := builder.Tx.SignPayload(a.Address, a.Key.Index, a.Signer); err != nil {
			return err
		}
	}

	// Proposer signs the payload
	if !proposer.Equals(payer) {
		if err := builder.Tx.SignPayload(proposer.Address, proposer.Key.Index, proposer.Signer); err != nil {
			return err
		}
	}

	// Payer signs the envelope
	if err := builder.Tx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer); err != nil {
		return err
	}

	transaction.ProposerAddress = flow_helpers.FormatAddress(proposer.Address)
	transaction.TransactionType = tType
	transaction.Actual = builder.Tx

	return nil
}

// Send the transaction to the network
func (t *Transaction) Send(ctx context.Context, fc *client.Client) error {
	err := fc.SendTransaction(ctx, *t.Actual)

	// Set TransactionId
	t.TransactionId = t.Actual.ID().Hex()

	return err
}

// Wait for the transaction to be sealed
func (t *Transaction) Wait(ctx context.Context, fc *client.Client, timeout time.Duration) error {
	result, err := flow_helpers.WaitForSeal(ctx, fc, t.Actual.ID(), timeout)
	if err != nil {
		return err
	}
	t.Result = result
	return nil
}

// Send the transaction to the network and wait for seal
func (t *Transaction) SendAndWait(ctx context.Context, fc *client.Client, timeout time.Duration) error {
	if err := t.Send(ctx, fc); err != nil {
		return err
	}

	// Wait for the transaction to be sealed
	if err := t.Wait(ctx, fc, timeout); err != nil {
		return err
	}

	return nil
}

func (t *Transaction) Hydrate(ctx context.Context, fc *client.Client) error {
	if t.Actual != nil {
		// Already hydrated
		return nil
	}

	actual, err := fc.GetTransaction(context.Background(), flow.HexToID(t.TransactionId))
	if err != nil {
		return err
	}

	t.Actual = actual

	return nil
}
