// Package aws provides functions for key and signer generation in AWS KMS.
package aws

import (
	"context"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

type kmsPubKeyMeta struct {
	Oid1 asn1.ObjectIdentifier
	Oid2 asn1.ObjectIdentifier
}

type kmsPubKey struct {
	Meta      kmsPubKeyMeta
	PublicKey asn1.BitString
}

// Generates an asymmetric signing & verification key (ECC_SECG_P256K1 / ECDSA_SHA_256) in AWS KMS
// and returns data required for account creation; a flow.AccountKey and a private key. The private
// key has the KMS key ARN as the value.
func Generate(cfg *configs.Config, ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *keys.Private, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := kms.NewFromConfig(awsCfg)

	// Create the new key in AWS KMS
	createKeyOutput, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		// CustomKeyStoreId: aws.String(""),                                                                         // TODO: Add support for custom key stores
		KeySpec:     types.KeySpecEccSecgP256k1,                                                             // TODO: Make key type configurable
		Description: aws.String(fmt.Sprintf("custodial account key for flow-wallet-api @ %s", cfg.ChainID)), // TODO: Add relevant meta data to description or tags
		KeyUsage:    types.KeyUsageTypeSignVerify,
		Tags: []types.Tag{
			{
				TagKey:   aws.String("ChainID"),
				TagValue: aws.String(string(cfg.ChainID)),
			},
			{
				TagKey:   aws.String("CreatedBy"),
				TagValue: aws.String("flow-wallet-api"), // TODO: Make this configurable, e.g. instance ID
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Get the public key from AWS KMS
	pbkOutput, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: createKeyOutput.KeyMetadata.KeyId})

	if err != nil {
		return nil, nil, err
	}

	var dest kmsPubKey

	// Decode the public key
	_, err = asn1.Unmarshal(pbkOutput.PublicKey, &dest)
	if err != nil {
		return nil, nil, err
	}

	// Convert the decoded public key into a PEM in string format so that the
	// DecodePublicKeyPEM from flow-go-sdk/crypto can be used
	pemStr := string(pem.EncodeToMemory(&pem.Block{Bytes: pbkOutput.PublicKey})[:])
	pbk, err := crypto.DecodePublicKeyPEM(crypto.ECDSA_secp256k1, pemStr) // TODO: Other key types
	if err != nil {
		return nil, nil, err
	}

	f := flow.NewAccountKey().
		SetPublicKey(pbk).
		SetHashAlgo(crypto.SHA3_256). // TODO: Read this from the key data
		SetWeight(weight)
	f.Index = keyIndex

	pk := &keys.Private{
		Index: keyIndex,
		Type:  keys.AccountKeyTypeAWSKMS,
		Value: *createKeyOutput.KeyMetadata.Arn,
	}

	return f, pk, nil
}

func Signer(ctx context.Context, address flow.Address, key keys.Private) (crypto.Signer, error) {
	s, err := SignerForKey(ctx, address, key)

	if err != nil {
		return nil, err
	}

	return s, nil
}

// Signer is a Google Cloud KMS implementation of crypto.Signer.
type AWSSigner struct {
	ctx     context.Context
	client  *kms.Client
	address flow.Address
	keyId   string
	hasher  crypto.Hasher
}

// SignerForKey returns a new Google Cloud KMS signer for an asymmetric key version.
func SignerForKey(
	ctx context.Context,
	address flow.Address,
	key keys.Private,
) (*AWSSigner, error) {
	a, err := arn.Parse(key.Value) // Get the key info by ARN in key.Value
	if err != nil {
		return nil, err
	}

	parts := strings.Split(a.Resource, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Resource in ARN; %s", a.Resource)
	}
	keyId := parts[1]

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := kms.NewFromConfig(cfg)

	// Get the public key from AWS KMS
	pbkOutput, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: aws.String(keyId)})

	if err != nil {
		return nil, err
	}

	var hashAlgo crypto.HashAlgorithm

	// Check that ECDSA_SHA_256 is available
	// TODO: Make hashing algorithm configurable
	for _, a := range pbkOutput.SigningAlgorithms {
		if a == types.SigningAlgorithmSpecEcdsaSha256 {
			hashAlgo = crypto.SHA3_256
			break
		}
	}

	if hashAlgo == crypto.UnknownHashAlgorithm {
		return nil, fmt.Errorf("unknown hash algorithm")
	}

	hasher, err := crypto.NewHasher(hashAlgo)
	if err != nil {
		return nil, fmt.Errorf("keys/aws: failed to instantiate hasher: %w", err)
	}

	return &AWSSigner{
		ctx:     ctx,
		client:  client,
		address: address,
		keyId:   keyId,
		hasher:  hasher,
	}, nil
}

// Sign signs the given message using the KMS signing key for this signer.
//
// Reference: https://docs.aws.amazon.com/kms/latest/APIReference/API_Sign.html
func (s *AWSSigner) Sign(message []byte) ([]byte, error) {
	digest := s.hasher.ComputeHash(message)

	sigOut, err := s.client.Sign(s.ctx, &kms.SignInput{
		KeyId:            &s.keyId,
		Message:          digest,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
	})

	if err != nil {
		return nil, fmt.Errorf("keys/aws: failed to sign: %w", err)
	}

	sig, err := parseSignature(sigOut.Signature)
	if err != nil {
		return nil, fmt.Errorf("keys/aws: failed to parse signature: %w", err)
	}

	return sig, nil
}

// Source: flow-go-sdk/crypto/cloudkms/signer.go

// ecCoupleComponentSize is size of a component in either (r,s) couple for an elliptical curve signature
// or (x,y) identifying a public key. Component size is needed for encoding couples comprised of variable length
// numbers to []byte encoding. They are not always the same length, so occasionally padding is required.
// Here's how one calculates the required length of each component:
// 		ECDSA_CurveBits = 256
// 		ecCoupleComponentSize := ECDSA_CurveBits / 8
// 		if ECDSA_CurveBits % 8 > 0 {
//			ecCoupleComponentSize++
// 		}
const ecCoupleComponentSize = 32

func parseSignature(signature []byte) ([]byte, error) {
	var parsedSig struct{ R, S *big.Int }
	if _, err := asn1.Unmarshal(signature, &parsedSig); err != nil {
		return nil, fmt.Errorf("asn1.Unmarshal: %w", err)
	}

	rBytes := parsedSig.R.Bytes()
	rBytesPadded := rightPad(rBytes, ecCoupleComponentSize)

	sBytes := parsedSig.S.Bytes()
	sBytesPadded := rightPad(sBytes, ecCoupleComponentSize)

	return append(rBytesPadded, sBytesPadded...), nil
}

// rightPad pads a byte slice with empty bytes (0x00) to the given length.
func rightPad(b []byte, length int) []byte {
	padded := make([]byte, length)
	copy(padded[length-len(b):], b)
	return padded
}
