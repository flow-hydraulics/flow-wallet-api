package transactions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
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
	Arguments          []CadenceArgument          `json:"arguments"`
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

	args, err := decodeCDCArguments(st.Arguments)
	if err != nil {
		return SignedTransactionJSONResponse{}, err
	}

	res.Code = string(st.Script)
	res.Arguments = args
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
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Events          []flow.Event   `gorm:"-"`
}

func (Transaction) TableName() string {
	return "transactions"
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

func New(
	transaction *Transaction,
	referenceBlockID flow.Identifier,
	builder *templates.TransactionBuilder,
	tType Type,
	proposer, payer keys.Authorizer,
	authorizers []keys.Authorizer) error {

	builder.Tx.
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address).
		SetGasLimit(maxGasLimit)

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
	transaction.TransactionId = builder.Tx.ID().Hex()

	return nil
}

func decodeCDCArguments(encodedArgs [][]byte) (args []CadenceArgument, err error) {
	for _, bs := range encodedArgs {
		var a CadenceArgument
		err := json.Unmarshal(bs, &a)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	return args, nil
}
