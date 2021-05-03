package keys

import (
	"github.com/eqlabs/flow-nft-wallet-service/data"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

const (
	ACCOUNT_KEY_TYPE_LOCAL      = "local"
	ACCOUNT_KEY_TYPE_GOOGLE_KMS = "google_kms"
)

type Manager interface {
	Generate(keyIndex int, weight int) (Wrapped, error)
	GenerateDefault() (Wrapped, error)
	Save(Key) (data.Key, error)
	Load(data.Key) (Key, error)
	AdminAuthorizer() (Authorizer, error)
	UserAuthorizer(address string) (Authorizer, error)
}

// "In flight" account key
type Key struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
	Value string `json:"-"`
}

type Authorizer struct {
	Address flow.Address
	Key     *flow.AccountKey
	Signer  crypto.Signer
}

type Wrapped struct {
	FlowKey    *flow.AccountKey
	AccountKey Key
}
