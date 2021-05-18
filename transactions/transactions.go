package transactions

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

type Transaction struct {
	ID             int                     `json:"-" gorm:"primaryKey"`
	AccountAddress string                  `json:"-" gorm:"index"`
	TransactionId  string                  `json:"transactionId" gorm:"index"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt          `json:"-" gorm:"index"`
	Code           string                  `json:"-" gorm:"-"`
	Arguments      []TransactionArg        `json:"-" gorm:"-"`
	tx             *flow.Transaction       `json:"-" gorm:"-"`
	result         *flow.TransactionResult `json:"-" gorm:"-"`
}

// https://docs.onflow.org/cadence/json-cadence-spec/
type TransactionArg interface{}

var EmptyTransaction Transaction = Transaction{}

// Send the transaction to the network
func (t *Transaction) Send(ctx context.Context, fc *client.Client) error {
	return fc.SendTransaction(ctx, *t.tx)
}

// Wait for the transaction to be sealed
func (t *Transaction) Wait(ctx context.Context, fc *client.Client) error {
	result, err := flow_helpers.WaitForSeal(ctx, fc, t.tx.ID())
	if err != nil {
		return err
	}
	t.result = result
	return nil
}

func New(
	referenceBlockID flow.Identifier,
	code string,
	args []TransactionArg,
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
	for _, arg := range args {
		jsonbytes, err := json.Marshal(arg)
		if err != nil {
			return &EmptyTransaction, err
		}
		tx.AddRawArgument(jsonbytes)
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
		Code:           code,
		Arguments:      args,
		tx:             tx,
	}, nil
}
