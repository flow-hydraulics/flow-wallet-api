package templates

import "github.com/onflow/flow-go-sdk"

type Raw struct {
	Code      string     `json:"code" gorm:"-"`
	Arguments []Argument `json:"arguments" gorm:"-"`
}

// https://docs.onflow.org/cadence/json-cadence-spec/
type Argument interface{}

type TransactionBuilder struct {
	Tx *flow.Transaction
}

func NewTransactionBuilder(raw Raw) (*TransactionBuilder, error) {
	t := flow.NewTransaction()
	t.SetScript([]byte(raw.Code))

	b := &TransactionBuilder{Tx: t}

	// Add arguments
	for _, a := range raw.Arguments {
		err := b.AddArgument(a)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (b *TransactionBuilder) AddArgument(a Argument) error {
	c, err := AsCadence(&a)
	if err != nil {
		return err
	}
	err = b.Tx.AddArgument(c)
	if err != nil {
		return err
	}
	return nil
}
