package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
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

type TransactionArg struct {
	Type  string
	Value string
}

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
	others []keys.Authorizer) (*Transaction, error) {

	arguments := make([]cadence.Value, len(args))

	for i, arg := range args {
		val, err := CadenceValue(arg)
		if err != nil {
			return &EmptyTransaction, err
		}
		arguments[i] = val
	}

	// Create Flow transaction
	// TODO: Gas limit?
	tx := flow.NewTransaction().
		SetScript([]byte(code)).
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address)

	// Add arguments
	for _, arg := range arguments {
		tx.AddRawArgument(jsoncdc.MustEncode(arg))
	}

	// Add authorizers
	tx.AddAuthorizer(payer.Address)
	for _, other := range others {
		tx.AddAuthorizer(other.Address)
	}

	// Sign the payload
	// TODO: support multiple keys per account
	for _, other := range others {
		err := tx.SignPayload(other.Address, other.Key.Index, other.Signer)
		if err != nil {
			return &EmptyTransaction, err
		}
	}

	// Sign the envelope
	// TODO: support multiple keys per account
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

func CadenceValue(arg TransactionArg) (cadence.Value, error) {
	switch strings.ToLower(arg.Type) {
	default:
		return cadence.Void{}, fmt.Errorf("unknown argument type %s", arg.Type)
	case "string":
		return cadence.NewString(arg.Value), nil
	case "address":
		return cadence.NewAddress(flow.HexToAddress(arg.Value)), nil
	case "ufix64":
		return cadence.NewUFix64(arg.Value)
	case "array":
		var arr []TransactionArg
		err := json.Unmarshal([]byte(arg.Value), &arr)
		if err != nil {
			return cadence.Void{}, err
		}
		values := make([]cadence.Value, len(arr))
		for i, arg := range arr {
			val, err := CadenceValue(arg)
			if err != nil {
				return cadence.Void{}, err
			}
			values[i] = val
		}
		return cadence.NewArray(values), nil
	}
}
