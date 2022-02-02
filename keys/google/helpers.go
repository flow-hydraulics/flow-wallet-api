package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"

	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// AsymKey creates a new asymmetric signing key in Google KMS and returns
// a cloudkms.Key (the "raw" result isn't needed)
func AsymKey(ctx context.Context, parent, id string) (*cloudkms.Key, error) {
	c, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Errorf("Failed to new kms client: %+v", err)
		return nil, err
	}

	// Close the client connection to avoid leaking goroutines
	defer c.Close()

	r := &kmspb.CreateCryptoKeyRequest{
		Parent:      parent,
		CryptoKeyId: id,
		CryptoKey: &kmspb.CryptoKey{
			Purpose: kmspb.CryptoKey_ASYMMETRIC_SIGN,
			VersionTemplate: &kmspb.CryptoKeyVersionTemplate{
				Algorithm: kmspb.CryptoKeyVersion_EC_SIGN_P256_SHA256,
			},
			// TODO: Set relevant labels at creation, update post-creation if necessary
			Labels: map[string]string{
				"service":         "flow-wallet-api",
				"account_address": "",
				"chain_id":        "",
				"environment":     "",
			},
		},
	}

	startCreate := time.Now()
	gk, err := c.CreateCryptoKey(ctx, r)
	if err != nil {
		log.Errorf("Failed to create crypto key version: %+v", err)
		return nil, err
	}
	log.Debugf("Create crypto key took %s", time.Since(startCreate))

	// Append cryptoKeyVersions so that we can utilize the KeyFromResourceID method
	k, err := cloudkms.KeyFromResourceID(fmt.Sprintf("%s/cryptoKeyVersions/1", gk.Name))
	if err != nil {
		return nil, err
	}

	keyVersion, err := c.GetCryptoKeyVersion(ctx, &kmspb.GetCryptoKeyVersionRequest{
		Name: k.ResourceID(),
	})
	if err != nil {
		log.Errorf("Failed to get crypto key version: %+v", err)
		return nil, err
	}
	if keyVersion.State != kmspb.CryptoKeyVersion_ENABLED {
		err := waitForKey(ctx, c, k.ResourceID())
		if err != nil {
			log.Warnf("Wait for key creation failed: %s %+v", k.ResourceID(), err)
		}
	}

	// Validate key name
	if !strings.HasPrefix(k.ResourceID(), gk.Name) {
		err := fmt.Errorf("WARNING: created Google KMS key name does not match the expected")
		return nil, err
	}

	return &k, nil
}

func waitForKey(ctx context.Context, client *kms.KeyManagementClient, keyVersionResourceID string) error {
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Minute,
		Factor: 5,
		Jitter: true,
	}

	for {
		keyVersion, err := client.GetCryptoKeyVersion(ctx, &kmspb.GetCryptoKeyVersionRequest{
			Name: keyVersionResourceID,
		})
		if err != nil {
			return err
		}
		if keyVersion.State == kmspb.CryptoKeyVersion_ENABLED {
			return nil
		}
		log.Debugf("Waiting for key creation: %s %s\n", keyVersion.State, keyVersionResourceID)
		time.Sleep(b.Duration())
	}
}
