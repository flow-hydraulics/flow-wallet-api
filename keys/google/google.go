// Package google provides functions for key and signer generation in Google KMS.
package google

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/google/uuid"
	"github.com/jpillora/backoff"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
	log "github.com/sirupsen/logrus"
)

func getPublicKey(ctx context.Context, c *cloudkms.Client, k *cloudkms.Key) (*crypto.PublicKey, *crypto.HashAlgorithm, *crypto.SignatureAlgorithm, error) {
	// Get the public key (using flow-go-sdk's cloudkms.Client)
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Minute,
		Factor: 5,
		Jitter: true,
	}

	deadline := time.Now().Add(60 * time.Second)

	var pub crypto.PublicKey
	var h crypto.HashAlgorithm
	var s crypto.SignatureAlgorithm
	var err error

	entry := log.WithFields(log.Fields{"keyId": k.KeyID})

	entry.Trace("Getting public key for KMS key")

	for {
		pub, h, err = c.GetPublicKey(ctx, *k)
		if pub != nil {
			s = pub.Algorithm()
			break
		}
		// non-retryable error
		if err != nil && !strings.Contains(err.Error(), "KEY_PENDING_GENERATION") {
			entry.WithFields(log.Fields{"err": err}).Error("failed to get public key")
			return nil, nil, nil, err
		}
		// key not generated yet, retry
		if err != nil && strings.Contains(err.Error(), "KEY_PENDING_GENERATION") {
			entry.Trace("KMS key is pending creation, will retry")
			continue
		}

		time.Sleep(b.Duration())

		if time.Now().After(deadline) {
			err = fmt.Errorf("timeout while trying to get public key")
			return nil, nil, nil, err
		}
	}

	return &pub, &h, &s, err
}

// Generate creates a new asymmetric signing & verification key in Google KMS
// and returns the required data to use the key with the Flow blockchain
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

	// TODO: The private key will be created, and ONLY the "get public key" part
	// should be retried
	pub, h, s, err := getPublicKey(ctx, c, k)
	if err != nil {
		log.WithFields(log.Fields{"keyId": k.KeyID, "err": err}).Error("failed to get public key for Google KMS key")
		return nil, nil, err
	}

	f := flow.NewAccountKey().
		SetPublicKey(*pub).
		SetHashAlgo(*h).
		SetWeight(weight)
	f.Index = keyIndex

	p := &keys.Private{
		Index:    keyIndex,
		Type:     keys.AccountKeyTypeGoogleKMS,
		Value:    k.ResourceID(),
		SignAlgo: *s,
		HashAlgo: *h,
	}

	return f, p, nil
}

// Signer creates a crypto.Signer for the given private key
// (KMS key resource name)
func Signer(ctx context.Context, key keys.Private) (crypto.Signer, error) {
	c, err := cloudkms.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	k, err := cloudkms.KeyFromResourceID(key.Value)
	if err != nil {
		return nil, err
	}

	s, err := c.SignerForKey(ctx, k)

	if err != nil {
		return nil, err
	}

	return s, nil
}
