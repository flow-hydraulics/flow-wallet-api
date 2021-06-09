package transactions

import (
	"context"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

type Transaction struct {
	ID              int                     `json:"-" gorm:"primaryKey"`
	PayerAddress    string                  `json:"-" gorm:"index"`
	TransactionId   string                  `json:"transactionId" gorm:"index"`
	TransactionType Type                    `json:"transactionType" gorm:"index"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt          `json:"-" gorm:"index"`
	Result          *flow.TransactionResult `json:"-" gorm:"-"`
	flowTx          *flow.Transaction       `json:"-" gorm:"-"`
}

func New(
	referenceBlockID flow.Identifier,
	builder *templates.TransactionBuilder,
	tType Type,
	proposer, payer keys.Authorizer,
	authorizers []keys.Authorizer) (*Transaction, error) {

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
	// TODO: support multiple keys per account?
	for _, a := range authorizers {
		err := builder.Tx.SignPayload(a.Address, a.Key.Index, a.Signer)
		if err != nil {
			return &Transaction{}, err
		}
	}

	// Payer signs the envelope
	// TODO: support multiple keys per account?
	err := builder.Tx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer)
	if err != nil {
		return &Transaction{}, err
	}

	return &Transaction{
		PayerAddress:    flow_helpers.FormatAddress(payer.Address),
		TransactionType: tType,
		flowTx:          builder.Tx,
	}, nil
}

// Send the transaction to the network
func (t *Transaction) Send(ctx context.Context, fc *client.Client) error {
	err := fc.SendTransaction(ctx, *t.flowTx)

	// Set TransactionId
	t.TransactionId = t.flowTx.ID().Hex()

	return err
}

// Wait for the transaction to be sealed
func (t *Transaction) Wait(ctx context.Context, fc *client.Client) error {
	result, err := flow_helpers.WaitForSeal(ctx, fc, t.flowTx.ID())
	if err != nil {
		return err
	}
	t.Result = result
	return nil
}

// Send the transaction to the network and wait for seal
func (t *Transaction) SendAndWait(ctx context.Context, fc *client.Client) error {
	err := t.Send(ctx, fc)
	if err != nil {
		return err
	}

	// Wait for the transaction to be sealed
	err = t.Wait(ctx, fc)
	if err != nil {
		return err
	}

	return err
}
