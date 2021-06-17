// Package google provides functions for key and signer generation in Google KMS.
package google

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/google/uuid"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

type Config struct {
	ProjectID  string `env:"GOOGLE_KMS_PROJECT_ID"`
	LocationID string `env:"GOOGLE_KMS_LOCATION_ID"`
	KeyRingID  string `env:"GOOGLE_KMS_KEYRING_ID"`
}

func Generate(ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *keys.Private, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, nil, err
	}

	u := uuid.New()

	// Create the new key in Google KMS
	k, err := AsymKey(
		ctx,
		fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", cfg.ProjectID, cfg.LocationID, cfg.KeyRingID),
		fmt.Sprintf("flow-wallet-account-key-%s", u.String()),
	)
	if err != nil {
		return nil, nil, err
	}

	c, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the public key (using flow-go-sdk's cloudkms.Client)
	pub, h, err := c.GetPublicKey(ctx, k)
	if err != nil {
		return nil, nil, err
	}

	f := flow.NewAccountKey().
		SetPublicKey(pub).
		SetHashAlgo(h).
		SetWeight(weight)
	f.Index = keyIndex

	p := &keys.Private{
		Index: keyIndex,
		Type:  keys.AccountKeyTypeGoogleKMS,
		Value: k.ResourceID(),
	}

	return f, p, nil
}

func Signer(ctx context.Context, address flow.Address, key keys.Private) (crypto.Signer, error) {
	c, err := cloudkms.NewClient(ctx)
	if err != nil {
		return &cloudkms.Signer{}, err
	}

	k, err := cloudkms.KeyFromResourceID(key.Value)
	if err != nil {
		return &cloudkms.Signer{}, err
	}

	s, err := c.SignerForKey(ctx, address, k)

	if err != nil {
		return &cloudkms.Signer{}, err
	}

	return s, nil
}
