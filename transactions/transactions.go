package transactions

import (
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"gorm.io/gorm"
)

const maxGasLimit = 9999

type SignedTransaction struct {
	flow.Transaction
}

// Signatures JSON HTTP response
type SignedTransactionJSONResponse struct {
	Code               string                     `json:"code"`
	Arguments          [][]byte                   `json:"arguments"`
	ReferenceBlockID   string                     `json:"referenceBlockId"`
	GasLimit           uint64                     `json:"gasLimit"`
	ProposalKey        ProposalKeyJSON            `json:"proposalKey"`
	Payer              string                     `json:"payer"`
	Authorizers        []string                   `json:"authorizers"`
	PayloadSignatures  []TransactionSignatureJSON `json:"payloadSignatures"`
	EnvelopeSignatures []TransactionSignatureJSON `json:"envelopeSignatures"`
}

type CadenceArgument interface{}

type ProposalKeyJSON struct {
	Address        string `json:"address"`
	KeyIndex       int    `json:"keyIndex"`
	SequenceNumber uint64 `json:"sequenceNumber"`
}

type TransactionSignatureJSON struct {
	Address   string `json:"address"`
	KeyIndex  int    `json:"keyIndex"`
	Signature string `json:"signature"`
}

func (st *SignedTransaction) ToJSONResponse() (SignedTransactionJSONResponse, error) {
	var res SignedTransactionJSONResponse

	res.Code = string(st.Script)
	res.Arguments = st.Arguments
	res.ReferenceBlockID = st.ReferenceBlockID.Hex()
	res.GasLimit = st.GasLimit
	res.ProposalKey = ProposalKeyJSON{
		Address:        st.ProposalKey.Address.Hex(),
		KeyIndex:       st.ProposalKey.KeyIndex,
		SequenceNumber: st.ProposalKey.SequenceNumber,
	}
	res.Payer = st.Payer.Hex()

	for _, a := range st.Authorizers {
		res.Authorizers = append(res.Authorizers, a.Hex())
	}

	for _, s := range st.PayloadSignatures {
		sig := TransactionSignatureJSON{
			Address:   s.Address.Hex(),
			KeyIndex:  s.KeyIndex,
			Signature: fmt.Sprintf("%x", s.Signature),
		}
		res.PayloadSignatures = append(res.PayloadSignatures, sig)
	}

	for _, s := range st.EnvelopeSignatures {
		sig := TransactionSignatureJSON{
			Address:   s.Address.Hex(),
			KeyIndex:  s.KeyIndex,
			Signature: fmt.Sprintf("%x", s.Signature),
		}
		res.EnvelopeSignatures = append(res.EnvelopeSignatures, sig)
	}

	return res, nil
}

// Transaction is the database model for all transactions.
type Transaction struct {
	TransactionId   string         `gorm:"column:transaction_id;primaryKey"`
	TransactionType Type           `gorm:"column:transaction_type;index"`
	ProposerAddress string         `gorm:"column:proposer_address;index"`
	FlowTransaction []byte         `gorm:"column:flow_transaction;type:bytes"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Events          []flow.Event   `gorm:"-"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// Transaction JSON HTTP request
type JSONRequest struct {
	Code      string     `json:"code"`
	Arguments []Argument `json:"arguments"`
}

// Transaction JSON HTTP response
type JSONResponse struct {
	TransactionId   string       `json:"transactionId"`
	TransactionType Type         `json:"transactionType"`
	Events          []flow.Event `json:"events,omitempty"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

func (t Transaction) ToJSONResponse() JSONResponse {
	return JSONResponse{
		TransactionId:   t.TransactionId,
		TransactionType: t.TransactionType,
		Events:          t.Events,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
