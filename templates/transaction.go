package templates

import (
	"fmt"

	"github.com/onflow/cadence"
	c_json "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

type Raw struct {
	Code      string     `json:"code" gorm:"-"`
	Arguments []Argument `json:"arguments" gorm:"-"`
}

// https://docs.onflow.org/cadence/json-cadence-spec/
type Argument interface{}

type TransactionBuilder struct {
	Tx *flow.Transaction
}

func NewBuilderFromRaw(raw Raw) (*TransactionBuilder, error) {
	t := flow.NewTransaction()
	t.SetScript([]byte(raw.Code))

	b := &TransactionBuilder{Tx: t}

	// Add arguments
	for _, a := range raw.Arguments {
		if err := b.AddArgument(a); err != nil {
			return nil, err
		}
	}

	return b, nil
}

func NewBuilderFromTx(tx *flow.Transaction) *TransactionBuilder {
	return &TransactionBuilder{Tx: tx}
}

func (b *TransactionBuilder) AddArgument(a Argument) error {
	c, err := AsCadence(&a)
	if err != nil {
		return err
	}
	if err := b.Tx.AddArgument(c); err != nil {
		return err
	}
	return nil
}

func (b *TransactionBuilder) GetArgument(index int) (cadence.Value, error) {
	if index < 0 || index >= len(b.Tx.Arguments) {
		return nil, fmt.Errorf("index out of bounds")
	}
	return c_json.Decode(b.Tx.Arguments[index])
}
