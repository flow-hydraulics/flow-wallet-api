package simple

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys/google"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"

	"github.com/google/uuid"
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

func (s *KeyManager) Generate(keyIndex int, weight int) (result keys.Wrapped, err error) {
	switch s.defaultKeyManager {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		seed := make([]byte, crypto.MinSeedLength)
		_, err := rand.Read(seed)
		if err != nil {
			break
		}

		privateKey, err := crypto.GeneratePrivateKey(s.signAlgo, seed)
		if err != nil {
			break
		}

		flowKey := flow.NewAccountKey().
			FromPrivateKey(privateKey).
			SetHashAlgo(s.hashAlgo).
			SetWeight(weight)
		flowKey.Index = keyIndex

		accountKey := keys.Key{
			Index: keyIndex,
			Type:  keys.ACCOUNT_KEY_TYPE_LOCAL,
			Value: privateKey.String(),
		}

		result.AccountKey = accountKey
		result.FlowKey = flowKey

	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		// TODO: Take this as a param / config instead
		gkmsConfig := GoogleKMSConfig{
			ProjectID:  os.Getenv("GOOGLE_KMS_PROJECT_ID"),
			LocationID: os.Getenv("GOOGLE_KMS_LOCATION_ID"),
			KeyRingID:  os.Getenv("GOOGLE_KMS_KEYRING_ID"),
		}

		ctx := context.Background()

		keyUUID := uuid.New()

		// Create the new key in Google KMS
		createdKey, err := google.AsyncKey(
			ctx,
			fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", gkmsConfig.ProjectID, gkmsConfig.LocationID, gkmsConfig.KeyRingID),
			fmt.Sprintf("flow-wallet-account-key-%s", keyUUID.String()),
		)
		if err != nil {
			break
		}

		client, err := cloudkms.NewClient(ctx)
		if err != nil {
			break
		}

		// Get the public key (using flow-go-sdk's cloudkms.Client)
		publicKey, hashAlgorithm, err := client.GetPublicKey(ctx, createdKey)
		if err != nil {
			break
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

		result.AccountKey = accountKey
		result.FlowKey = flowKey

	default:
		err = fmt.Errorf("keyStore.Generate() not implmented for %s", s.defaultKeyManager)
	}
	return
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
