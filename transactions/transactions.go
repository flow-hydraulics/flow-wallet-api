package transactions

import (
	"context"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

// https://docs.onflow.org/cadence/json-cadence-spec/
type Argument interface{}

type Script struct {
	Code      string     `json:"code" gorm:"-"`
	Arguments []Argument `json:"arguments" gorm:"-"`
}

type Transaction struct {
	Script
	ID             int                     `json:"-" gorm:"primaryKey"`
	AccountAddress string                  `json:"-" gorm:"index"`
	TransactionId  string                  `json:"transactionId" gorm:"index"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt          `json:"-" gorm:"index"`
	Result         *flow.TransactionResult `json:"-" gorm:"-"`
	flowTx         *flow.Transaction       `json:"-" gorm:"-"`
}

var EmptyTransaction Transaction = Transaction{}

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

func (t *Transaction) SetReferenceBlockID(id flow.Identifier) {
	t.flowTx.SetReferenceBlockID(id)
}

func New(
	referenceBlockID flow.Identifier,
	code string,
	args []Argument,
	proposer, payer keys.Authorizer,
	authorizers []keys.Authorizer) (*Transaction, error) {

	// Create Flow transaction
	// TODO: Gas limit?
	tx := flow.NewTransaction().
		SetScript([]byte(code)).
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address)

	// Add arguments
	for _, a := range args {
		c, err := AsCadence(&a)
		if err != nil {
			return &EmptyTransaction, err
		}
		tx.AddArgument(c)
	}

	// Add authorizers
	for _, a := range authorizers {
		tx.AddAuthorizer(a.Address)
	}

	// Authorizers sign the payload
	// TODO: support multiple keys per account?
	for _, a := range authorizers {
		err := tx.SignPayload(a.Address, a.Key.Index, a.Signer)
		if err != nil {
			return &EmptyTransaction, err
		}
	}

	// Payer signs the envelope
	// TODO: support multiple keys per account?
	err := tx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer)
	if err != nil {
		return &EmptyTransaction, err
	}

	return &Transaction{
		AccountAddress: payer.Address.Hex(),
		Script: Script{
			Code:      code,
			Arguments: args,
		},
		flowTx: tx,
	}, nil
}
