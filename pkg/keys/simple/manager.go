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

type KeyManager struct {
	db                data.Store
	fc                *client.Client
	serviceAddress    string
	serviceKey        keys.Key
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

func NewKeyManager(
	db data.Store,
	fc *client.Client,
	serviceAddress string,
	serviceKey keys.Key,
	defaultKeyManager string,
	encryptionKey string,
) (*KeyManager, error) {
	return &KeyManager{
		db,
		fc,
		serviceAddress,
		serviceKey,
		defaultKeyManager,
		NewCrypter([]byte(encryptionKey)),
		crypto.ECDSA_P256, // TODO: config
		crypto.SHA3_256,   // TODO: config
	}, nil
}

func (s *KeyManager) Generate(keyIndex int, weight int) (keys.Wrapped, error) {
	switch s.defaultKeyManager {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		seed := make([]byte, crypto.MinSeedLength)
		_, err := rand.Read(seed)
		if err != nil {
			return keys.Wrapped{}, err
		}
		privateKey, err := crypto.GeneratePrivateKey(s.signAlgo, seed)
		if err != nil {
			return keys.Wrapped{}, err
		}

		flowKey := flow.NewAccountKey().
			FromPrivateKey(privateKey).
			SetHashAlgo(s.hashAlgo).
			SetWeight(weight)
		flowKey.Index = keyIndex

		encPrivKey, err := s.crypter.Encrypt([]byte(privateKey.String()))
		if err != nil {
			return keys.Wrapped{}, err
		}

		accountKey := keys.Key{
			Index: keyIndex,
			Type:  keys.ACCOUNT_KEY_TYPE_LOCAL,
			Value: string(encPrivKey),
		}
		return keys.Wrapped{FlowKey: flowKey, AccountKey: accountKey}, nil
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		// TODO: Take this as a param / config instead
		gkmsConfig := GoogleKMSConfig{
			ProjectID:  os.Getenv("GOOGLE_KMS_PROJECT_ID"),
			LocationID: os.Getenv("GOOGLE_KMS_LOCATION_ID"),
			KeyRingID:  os.Getenv("GOOGLE_KMS_KEYRING_ID"),
		}
		keyUUID := uuid.New()

		ctx := context.Background()

		// Create the new key in Google KMS
		createdKey, err := createKeyAsymmetricSigningKey(
			ctx,
			fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", gkmsConfig.ProjectID, gkmsConfig.LocationID, gkmsConfig.KeyRingID),
			fmt.Sprintf("flow-wallet-account-key-%s", keyUUID.String()),
		)
		if err != nil {
			fmt.Println(err)
			return keys.Wrapped{}, err
		}

		client, err := cloudkms.NewClient(ctx)
		if err != nil {
			fmt.Println(err)
			return keys.Wrapped{}, err
		}

		// Get the public key (using flow-go-sdk's cloudkms.Client)
		publicKey, hashAlgorithm, err := client.GetPublicKey(ctx, createdKey)
		if err != nil {
			fmt.Println(err)
			return keys.Wrapped{}, err
		}

		accountKey := keys.Key{
			Index: keyIndex,
			Type:  keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS,
			Value: createdKey.ResourceID(),
		}

		flowKey := flow.NewAccountKey().
			SetPublicKey(publicKey).
			SetHashAlgo(hashAlgorithm).
			SetWeight(weight)
		flowKey.Index = keyIndex

		return keys.Wrapped{FlowKey: flowKey, AccountKey: accountKey}, nil
	default:
		return keys.Wrapped{}, fmt.Errorf("keyStore.Generate() not implmented for %s", s.defaultKeyManager)
	}
}

func (s *KeyManager) Save(key keys.Key) (result data.Key, err error) {
	encValue, err := s.crypter.Encrypt([]byte(key.Value))
	if err != nil {
		return
	}
	result.Index = key.Index
	result.Type = key.Type
	result.Value = encValue
	return
}

func (s *KeyManager) Load(key data.Key) (result keys.Key, err error) {
	decValue, err := s.crypter.Decrypt([]byte(key.Value))
	if err != nil {
		return
	}
	result.Index = key.Index
	result.Type = key.Type
	result.Value = string(decValue)
	return
}

func (s *KeyManager) AdminAuthorizer() (keys.Authorizer, error) {
	return s.MakeAuthorizer(s.serviceAddress)
}

func (s *KeyManager) UserAuthorizer(address string) (keys.Authorizer, error) {
	return s.MakeAuthorizer(address)
}

func (s *KeyManager) MakeAuthorizer(address string) (authorizer keys.Authorizer, err error) {
	var accountKey keys.Key
	ctx := context.Background()

	authorizer.Address = flow.HexToAddress(address)

	if address == s.serviceAddress {
		accountKey = s.serviceKey
	} else {
		dbKey, err := s.db.AccountKey(address, 0)
		if err != nil {
			return authorizer, err
		}
		accountKey, err = s.Load(dbKey)
		if err != nil {
			return authorizer, err
		}
	}

	flowAcc, err := s.fc.GetAccount(ctx, flow.HexToAddress(address))
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
