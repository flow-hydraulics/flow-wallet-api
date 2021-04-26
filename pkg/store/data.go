package store

import "github.com/onflow/flow-go-sdk"

type Account struct {
	Address flow.Address `json:"address"`
}

type AccountKey struct {
	AccountAddress flow.Address `json:"address"`
	Index          int          `json:"index"`
	Type           string       `json:"type"`  // local, google_kms
	Value          string       `json:"value"` // local: private key, google_kms: resource id
}

const (
	ACCOUNT_KEY_TYPE_LOCAL      = "local"
	ACCOUNT_KEY_TYPE_GOOGLE_KMS = "google_kms"
)
