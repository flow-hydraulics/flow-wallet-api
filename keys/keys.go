// Package keys provides key management functions.
package keys

import (
	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

const (
	ACCOUNT_KEY_TYPE_LOCAL      = "local"
	ACCOUNT_KEY_TYPE_GOOGLE_KMS = "google_kms"
)

// Manager provides the functions needed for key management.
type Manager interface {
	// Generate generates a new Key using provided key index and weight.
	Generate(keyIndex int, weight int) (Wrapped, error)
	// GenerateDefault generates a new Key using application defaults.
	GenerateDefault() (Wrapped, error)
	// Save is responsible for converting an "in flight" key to a storable key.
	Save(Key) (data.Key, error)
	// Load is responsible for converting a storable key to an "in flight" key.
	Load(data.Key) (Key, error)
	// AdminAuthorizer returns an Authorizer for the applications admin account.
	AdminAuthorizer() (Authorizer, error)
	// UserAuthorizer returns an Authorizer for the given address.
	UserAuthorizer(address string) (Authorizer, error)
}

// Key is an "in flight" account key meaning its Value should be the actual
// private key or resource id (unencrypted).
type Key struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
	Value string `json:"-"`
}

// Authorizer groups the necessary items for transaction signing.
type Authorizer struct {
	Address flow.Address
	Key     *flow.AccountKey
	Signer  crypto.Signer
}

// Wrapped simply provides a way to pass a flow.AccountKey and the corresponding Key together.
type Wrapped struct {
	FlowKey    *flow.AccountKey
	AccountKey Key
}
