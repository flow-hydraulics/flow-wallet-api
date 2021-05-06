// Package google provides functions for key and signer generation in Google KMS.
package google

import (
	"context"
	"fmt"

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

func Generate(
	projectId, locationId, KeyRingId string,
	keyIndex, weight int,
) (result keys.Wrapped, err error) {
	ctx := context.Background()

	keyUUID := uuid.New()

	// Create the new key in Google KMS
	createdKey, err := AsymKey(
		ctx,
		fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", projectId, locationId, KeyRingId),
		fmt.Sprintf("flow-wallet-account-key-%s", keyUUID.String()),
	)
	if err != nil {
		return
	}

	client, err := cloudkms.NewClient(ctx)
	if err != nil {
		return
	}

	// Get the public key (using flow-go-sdk's cloudkms.Client)
	publicKey, hashAlgorithm, err := client.GetPublicKey(ctx, createdKey)
	if err != nil {
		return
	}

	key := keys.Key{
		Index: keyIndex,
		Type:  keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS,
		Value: createdKey.ResourceID(),
	}

	flowKey := flow.NewAccountKey().
		SetPublicKey(publicKey).
		SetHashAlgo(hashAlgorithm).
		SetWeight(weight)
	flowKey.Index = keyIndex

	result.AccountKey = key
	result.FlowKey = flowKey

	return
}

func Signer(
	ctx context.Context,
	address string,
	key keys.Key,
) (result crypto.Signer, err error) {
	kmsClient, err := cloudkms.NewClient(ctx)
	if err != nil {
		return
	}

	kmsKey, err := cloudkms.KeyFromResourceID(key.Value)
	if err != nil {
		return
	}

	result, err = kmsClient.SignerForKey(
		ctx,
		flow.HexToAddress(address),
		kmsKey,
	)

	return
}
