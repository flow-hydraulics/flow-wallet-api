package simple

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/google/uuid"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type KeyStore struct {
	db                data.Store
	serviceAcct       data.AccountKey
	defaultKeyManager string
	crypter           *SymmetricCrypter
	signAlgo          crypto.SignatureAlgorithm
	hashAlgo          crypto.HashAlgorithm
}

type GoogleKMSConfig struct {
	ProjectID  string
	LocationID string
	KeyRingID  string
}

func NewKeyStore(
	db data.Store,
	serviceAcct data.AccountKey,
	defaultKeyManager string,
	encryptionKey string,
) (*KeyStore, error) {
	return &KeyStore{
		db,
		serviceAcct,
		defaultKeyManager,
		NewCrypter([]byte(encryptionKey)),
		crypto.ECDSA_P256, // TODO: config
		crypto.SHA3_256,   // TODO: config
	}, nil
}

func (s *KeyStore) Generate(ctx context.Context, keyIndex int, weight int) (keys.NewKeyWrapper, error) {
	switch s.defaultKeyManager {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		seed := make([]byte, crypto.MinSeedLength)
		_, err := rand.Read(seed)
		if err != nil {
			return keys.NewKeyWrapper{}, err
		}
		privateKey, err := crypto.GeneratePrivateKey(s.signAlgo, seed)
		if err != nil {
			return keys.NewKeyWrapper{}, err
		}

		flowKey := flow.NewAccountKey().
			FromPrivateKey(privateKey).
			SetHashAlgo(s.hashAlgo).
			SetWeight(weight)
		flowKey.Index = keyIndex

		accountKey := data.AccountKey{
			Index: keyIndex,
			Type:  keys.ACCOUNT_KEY_TYPE_LOCAL,
			Value: privateKey.String(),
		}
		return keys.NewKeyWrapper{FlowKey: flowKey, AccountKey: accountKey}, nil
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		// TODO: Take this as a param / config instead
		gkmsConfig := GoogleKMSConfig{
			ProjectID:  os.Getenv("GOOGLE_KMS_PROJECT_ID"),
			LocationID: os.Getenv("GOOGLE_KMS_LOCATION_ID"),
			KeyRingID:  os.Getenv("GOOGLE_KMS_KEYRING_ID"),
		}
		keyUUID := uuid.New()

		// Create the new key in Google KMS
		createdKey, err := createKeyAsymmetricSign(
			ctx,
			fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", gkmsConfig.ProjectID, gkmsConfig.LocationID, gkmsConfig.KeyRingID),
			fmt.Sprintf("flow-wallet-account-key-%s", keyUUID.String()),
		)
		if err != nil {
			fmt.Println(err)
			return keys.NewKeyWrapper{}, err
		}

		client, err := cloudkms.NewClient(ctx)
		if err != nil {
			fmt.Println(err)
			return keys.NewKeyWrapper{}, err
		}

		// Get the public key (using flow-go-sdk's cloudkms.Client)
		publicKey, hashAlgorithm, err := client.GetPublicKey(ctx, createdKey)
		if err != nil {
			fmt.Println(err)
			return keys.NewKeyWrapper{}, err
		}

		accountKey := data.AccountKey{
			Index: keyIndex,
			Type:  keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS,
			Value: createdKey.ResourceID(),
		}

		flowKey := flow.NewAccountKey().
			SetPublicKey(publicKey).
			SetHashAlgo(hashAlgorithm).
			SetWeight(weight)
		flowKey.Index = keyIndex

		return keys.NewKeyWrapper{FlowKey: flowKey, AccountKey: accountKey}, nil
	default:
		return keys.NewKeyWrapper{}, fmt.Errorf("keyStore.Generate() not implmented for %s", s.defaultKeyManager)
	}
}

func (s *KeyStore) Save(key data.AccountKey) error {
	switch key.Type {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		encValue, err := s.crypter.Encrypt([]byte(key.Value))
		if err != nil {
			return err
		}
		key.Value = string(encValue)
		err = s.db.InsertAccountKey(key)
		return err
	default:
		// TODO: google_kms
		return fmt.Errorf("keyStore.Save() not implmented for %s", s.defaultKeyManager)
	}
}

func (s *KeyStore) Delete(address string, keyIndex int) error {
	panic("not implemented") // TODO: implement
}

func (s *KeyStore) ServiceAuthorizer(ctx context.Context, fc *client.Client) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, fc, s.serviceAcct.AccountAddress)
}

func (s *KeyStore) AccountAuthorizer(ctx context.Context, fc *client.Client, address string) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, fc, address)
}

func (s *KeyStore) MakeAuthorizer(ctx context.Context, fc *client.Client, address string) (keys.Authorizer, error) {
	var (
		accountKey data.AccountKey
		authorizer keys.Authorizer = keys.Authorizer{}
		err        error
	)

	authorizer.Address = flow.HexToAddress(address)

	if address == s.serviceAcct.AccountAddress {
		accountKey = s.serviceAcct
	} else {
		accountKey, err = s.db.AccountKey(address)
		if err != nil {
			return authorizer, err
		}
		decValue, err := s.crypter.Decrypt([]byte(accountKey.Value))
		if err != nil {
			return authorizer, err
		}
		accountKey.Value = string(decValue)
	}

	flowAcc, err := fc.GetAccount(ctx, flow.HexToAddress(address))
	if err != nil {
		return authorizer, err
	}

	authorizer.Key = flowAcc.Keys[accountKey.Index]

	// TODO: Decide whether we want to allow this kind of flexibility
	// or should we just panic if `accountKey.Type` != `s.defaultKeyManager`
	switch accountKey.Type {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		pk, err := crypto.DecodePrivateKeyHex(s.signAlgo, accountKey.Value)
		if err != nil {
			return authorizer, err
		}
		authorizer.Signer = crypto.NewInMemorySigner(pk, s.hashAlgo)
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		kmsClient, err := cloudkms.NewClient(ctx)
		if err != nil {
			return authorizer, err
		}

		kmsKey, err := cloudkms.KeyFromResourceID(accountKey.Value)
		if err != nil {
			return authorizer, err
		}

		sig, err := kmsClient.SignerForKey(
			ctx,
			flow.HexToAddress(address),
			kmsKey,
		)
		if err != nil {
			return authorizer, err
		}
		authorizer.Signer = sig
	default:
		return authorizer,
			fmt.Errorf("accountKey.Type not recognised: %s", accountKey.Type)
	}

	return authorizer, nil
}

// Creates a new asymmetric signing key in Google KMS and returns a cloudkms.Key (the "raw" result isn't needed)
func createKeyAsymmetricSigningKey(ctx context.Context, parent, id string) (createdKey cloudkms.Key, err error) {
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
				"service":         "flow-nft-wallet-service",
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
		fmt.Println("Could not create cloudkms.Key from ResourceId:", googleKey.Name)
		return
	}

	// Validate key name
	if !strings.HasPrefix(createdKey.ResourceID(), googleKey.Name) {
		fmt.Println("WARNING: created Google KMS key name does not match the expected", createdKey.ResourceID(), " vs ", googleKey.Name)
		// TODO: Handle scenario
	}

	return
}
