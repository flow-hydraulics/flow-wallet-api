package google

import (
	"context"
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// AsymKey creates a new asymmetric signing key in Google KMS and returns a cloudkms.Key (the "raw" result isn't needed)
func AsymKey(ctx context.Context, parent, id string) (createdKey cloudkms.Key, err error) {
	kmsClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return
	}

	req := &kmspb.CreateCryptoKeyRequest{
		Parent:      parent,
		CryptoKeyId: id,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ASYMMETRIC_SIGN,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			},
			// TODO: Set relevant labels at creation, update post-creation if necessary
			Labels: map[string]string{
				"service":         "flow-wallet-service",
				"account_address": "",
				"chain_id":        "",
				"environment":     "development",
			},
		},
	}

	googleKey, err := kmsClient.CreateCryptoKey(ctx, req)
	if err != nil {
		return
	}

	// Append cryptoKeyVersions so that we can utilize the KeyFromResourceID method
	createdKey, err = cloudkms.KeyFromResourceID(fmt.Sprintf("%s/cryptoKeyVersions/1", googleKey.Name))
	if err != nil {
		return
	}

	// Validate key name
	if !strings.HasPrefix(createdKey.ResourceID(), googleKey.Name) {
		err = fmt.Errorf("WARNING: created Google KMS key name does not match the expected")
		return
	}

	return
}
