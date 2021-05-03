package simple

import (
	"context"
	"fmt"
	"log"

	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/keys/encryption"
	"github.com/eqlabs/flow-wallet-service/keys/google"
	"github.com/eqlabs/flow-wallet-service/keys/local"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

type Datastore interface {
	AccountKey(address string, index int) (data.Key, error)
}

type KeyManager struct {
	l               *log.Logger
	db              Datastore
	fc              *client.Client
	crypter         *encryption.SymmetricCrypter
	signAlgo        crypto.SignatureAlgorithm
	hashAlgo        crypto.HashAlgorithm
	adminAccountKey keys.Key
}

func NewKeyManager(l *log.Logger, db Datastore, fc *client.Client) (result *KeyManager, err error) {
	adminAccountKey := keys.Key{
		Index: cfg.AdminAccountKeyIndex,
		Type:  cfg.AdminAccountKeyType,
		Value: cfg.AdminAccountKeyValue,
	}

	crypter := encryption.NewCrypter([]byte(cfg.EncryptionKey))

	result = &KeyManager{
		l,
		db,
		fc,
		crypter,
		crypto.ECDSA_P256, // TODO: config
		crypto.SHA3_256,   // TODO: config
		adminAccountKey,
	}

	return
}

func (s *KeyManager) GenerateDefault() (keys.Wrapped, error) {
	return s.Generate(cfg.DefaultKeyIndex, cfg.DefaultKeyWeight)
}

func (s *KeyManager) Generate(keyIndex, weight int) (result keys.Wrapped, err error) {
	switch cfg.DefaultKeyStorage {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		result, err = local.Generate(
			s.signAlgo,
			s.hashAlgo,
			keyIndex,
			weight,
		)
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		result, err = google.Generate(
			googleCfg.ProjectID,
			googleCfg.LocationID,
			googleCfg.KeyRingID,
			keyIndex,
			weight,
		)
	default:
		err = fmt.Errorf("keyStore.Generate() not implmented for %s", cfg.DefaultKeyStorage)
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
	return s.MakeAuthorizer(cfg.AdminAccountAddress)
}

func (s *KeyManager) UserAuthorizer(address string) (keys.Authorizer, error) {
	return s.MakeAuthorizer(address)
}

func (s *KeyManager) MakeAuthorizer(address string) (result keys.Authorizer, err error) {
	var key keys.Key
	ctx := context.Background()

	result.Address = flow.HexToAddress(address)

	if address == cfg.AdminAccountAddress {
		key = s.adminAccountKey
	} else {
		var rawKey data.Key
		rawKey, err = s.db.AccountKey(address, 0)
		if err != nil {
			return
		}
		key, err = s.Load(rawKey)
		if err != nil {
			return
		}
	}

	flowAcc, err := s.fc.GetAccount(ctx, flow.HexToAddress(address))
	if err != nil {
		return
	}

	result.Key = flowAcc.Keys[key.Index]

	// TODO: Decide whether we want to allow this kind of flexibility
	// or should we just panic if `key.Type` != `s.defaultKeyManager`
	switch key.Type {
	case keys.ACCOUNT_KEY_TYPE_LOCAL:
		signer, err := local.Signer(s.signAlgo, s.hashAlgo, key)
		if err != nil {
			break
		}
		result.Signer = signer
	case keys.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
		signer, err := google.Signer(ctx, address, key)
		if err != nil {
			break
		}
		result.Signer = signer
	default:
		err = fmt.Errorf("key.Type not recognised: %s", key.Type)
	}

	return
}
