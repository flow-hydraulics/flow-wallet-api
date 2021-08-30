// Package google provides functions for key and signer generation in Google KMS.
package google

import (
	"context"
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/google/uuid"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

func Generate(cfg *configs.Config, ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *keys.Private, error) {
	u := uuid.New()

	// Create the new key in Google KMS
	k, err := AsymKey(
		ctx,
		fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", cfg.GoogleKMSProjectID, cfg.GoogleKMSLocationID, cfg.GoogleKMSKeyRingID),
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
	pub, h, err := c.GetPublicKey(ctx, *k)
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
