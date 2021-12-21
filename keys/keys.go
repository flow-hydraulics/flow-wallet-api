// Package keys provides key management functions.
package keys

import (
	"context"
	"errors"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"gorm.io/gorm"
)

const (
	AccountKeyTypeLocal     = "local"
	AccountKeyTypeGoogleKMS = "google_kms"
	AccountKeyTypeAWSKMS    = "aws_kms"
)

var ErrAdminProposalKeyCountMismatch = errors.New("admin-proposal-key count mismatch")

// Manager provides the functions needed for key management.
type Manager interface {
	// Generate generates a new Key using provided key index and weight.
	Generate(ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *Private, error)
	// GenerateDefault generates a new Key using application defaults.
	GenerateDefault(context.Context) (*flow.AccountKey, *Private, error)
	// Save is responsible for converting an "in flight" key to a storable key.
	Save(Private) (Storable, error)
	// Load is responsible for converting a storable key to an "in flight" key.
	Load(Storable) (Private, error)
	// AdminAuthorizer returns an Authorizer for the applications admin account.
	AdminAuthorizer(context.Context) (Authorizer, error)
	// UserAuthorizer returns an Authorizer for the given address.
	UserAuthorizer(ctx context.Context, address flow.Address) (Authorizer, error)
	// CheckAdminProposalKeyCount checks if admin proposal keys have been correctly initiated (counts match).
	CheckAdminProposalKeyCount(ctx context.Context) error
	// InitAdminProposalKeys will init the admin proposal keys in the database
	// and return current count.
	InitAdminProposalKeys(ctx context.Context) (uint16, error)
	// AdminProposalKey returns Authorizer to be used as proposer.
	AdminProposalKey(ctx context.Context) (Authorizer, error)
}

// Storable struct represents a storable account private key.
// Storable.Value is an encrypted byte representation of
// the actual private key when using local key management
// or resource id when using a remote key management system (e.g. Google KMS).
type Storable struct {
	ID             int            `json:"-" gorm:"primaryKey"`
	AccountAddress string         `json:"-" gorm:"index"`
	Index          int            `json:"index" gorm:"index"`
	Type           string         `json:"type"`
	Value          []byte         `json:"-"`
	PublicKey      string         `json:"publicKey"`
	SignAlgo       string         `json:"signAlgo"`
	HashAlgo       string         `json:"hashAlgo"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// Rename the database table to improve database readability
func (Storable) TableName() string {
	return "storable_keys"
}

type ProposalKey struct {
	ID        int `json:"-" gorm:"primaryKey"`
	KeyIndex  int `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProposalKey) TableName() string {
	return "proposal_keys"
}

// Private is an "in flight" account private key meaning its Value should be the actual
// private key or resource id (unencrypted).
type Private struct {
	Index    int                       `json:"index"`
	Type     string                    `json:"type"`
	Value    string                    `json:"-"`
	SignAlgo crypto.SignatureAlgorithm `json:"-"`
	HashAlgo crypto.HashAlgorithm      `json:"-"`
}

// Authorizer groups the necessary items for transaction signing.
type Authorizer struct {
	Address flow.Address
	Key     *flow.AccountKey
	Signer  crypto.Signer
}

func (a *Authorizer) Equals(t Authorizer) bool {
	return a.Address.Hex() == t.Address.Hex() && a.Key.Index == t.Key.Index
}
