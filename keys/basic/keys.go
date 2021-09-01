// Package basic provides straightforward implementation for key management.
package basic

import (
	"context"
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/keys/encryption"
	"github.com/flow-hydraulics/flow-wallet-api/keys/google"
	"github.com/flow-hydraulics/flow-wallet-api/keys/local"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

type KeyManager struct {
	store           keys.Store
	fc              *client.Client
	crypter         encryption.Crypter
	adminAccountKey keys.Private
	cfg             *configs.Config
}

// NewKeyManager initiates a new key manager.
// It uses encryption.AESCrypter to encrypt and decrypt the keys.
func NewKeyManager(cfg *configs.Config, store keys.Store, fc *client.Client) *KeyManager {
	// TODO(latenssi): safeguard against nil config?

	if cfg.DefaultKeyWeight < 0 {
		cfg.DefaultKeyWeight = flow.AccountKeyWeightThreshold
	}

	adminAccountKey := keys.Private{
		Index:    cfg.AdminKeyIndex,
		Type:     cfg.AdminKeyType,
		Value:    cfg.AdminPrivateKey,
		SignAlgo: crypto.StringToSignatureAlgorithm(cfg.DefaultSignAlgo),
		HashAlgo: crypto.StringToHashAlgorithm(cfg.DefaultHashAlgo),
	}

	crypter := encryption.NewAESCrypter([]byte(cfg.EncryptionKey))

	return &KeyManager{
		store,
		fc,
		crypter,
		adminAccountKey,
		cfg,
	}
}

func (s *KeyManager) Generate(ctx context.Context, keyIndex, weight int) (*flow.AccountKey, *keys.Private, error) {
	switch s.cfg.DefaultKeyType {
	default:
		return nil, nil, fmt.Errorf("keyStore.Generate() not implmented for %s", s.cfg.DefaultKeyType)
	case keys.AccountKeyTypeLocal:
		return local.Generate(
			keyIndex, weight,
			crypto.StringToSignatureAlgorithm(s.cfg.DefaultSignAlgo),
			crypto.StringToHashAlgorithm(s.cfg.DefaultHashAlgo))
	case keys.AccountKeyTypeGoogleKMS:
		return google.Generate(s.cfg, ctx, keyIndex, weight)
	}
}

func (s *KeyManager) GenerateDefault(ctx context.Context) (*flow.AccountKey, *keys.Private, error) {
	return s.Generate(ctx, s.cfg.DefaultKeyIndex, s.cfg.DefaultKeyWeight)
}

func (s *KeyManager) Save(key keys.Private) (keys.Storable, error) {
	encValue, err := s.crypter.Encrypt([]byte(key.Value))
	if err != nil {
		return keys.Storable{}, err
	}
	return keys.Storable{
		Index:    key.Index,
		Type:     key.Type,
		Value:    encValue,
		SignAlgo: key.SignAlgo.String(),
		HashAlgo: key.HashAlgo.String(),
	}, nil
}

func (s *KeyManager) Load(key keys.Storable) (keys.Private, error) {
	decValue, err := s.crypter.Decrypt([]byte(key.Value))
	if err != nil {
		return keys.Private{}, err
	}
	return keys.Private{
		Index:    key.Index,
		Type:     key.Type,
		Value:    string(decValue),
		SignAlgo: crypto.StringToSignatureAlgorithm(key.SignAlgo),
		HashAlgo: crypto.StringToHashAlgorithm(key.HashAlgo),
	}, nil
}

func (s *KeyManager) AdminAuthorizer(ctx context.Context) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, flow.HexToAddress(s.cfg.AdminAddress))
}

func (s *KeyManager) UserAuthorizer(ctx context.Context, address flow.Address) (keys.Authorizer, error) {
	return s.MakeAuthorizer(ctx, address)
}

func (s *KeyManager) MakeAuthorizer(ctx context.Context, address flow.Address) (keys.Authorizer, error) {
	var k keys.Private

	if address == flow.HexToAddress(s.cfg.AdminAddress) {
		k = s.adminAccountKey
	} else {
		// Get the "least recently used" key for this address
		sk, err := s.store.AccountKey(flow_helpers.FormatAddress(address))
		if err != nil {
			return keys.Authorizer{}, err
		}
		k, err = s.Load(sk)
		if err != nil {
			return keys.Authorizer{}, err
		}
	}

	acc, err := s.fc.GetAccount(ctx, address)
	if err != nil {
		return keys.Authorizer{}, err
	}

	sig, err := signerForKey(ctx, address, k)
	if err != nil {
		return keys.Authorizer{}, err
	}

	return keys.Authorizer{
		Address: address,
		Key:     acc.Keys[k.Index],
		Signer:  sig,
	}, nil
}

func (s *KeyManager) InitAdminProposalKeys(ctx context.Context) (uint16, error) {
	adminAddress := flow.HexToAddress(s.cfg.AdminAddress)

	adminAccount, err := s.fc.GetAccount(ctx, adminAddress)
	if err != nil {
		return 0, err
	}

	err = s.store.DeleteAllProposalKeys()
	if err != nil {
		return 0, err
	}

	var count uint16
	for _, k := range adminAccount.Keys {
		if !k.Revoked {
			err = s.store.InsertProposalKey(keys.ProposalKey{
				KeyIndex: k.Index,
			})
			if err != nil {
				return count, err
			}
			count += 1
		}
	}

	return count, nil
}

func (s *KeyManager) AdminProposalKey(ctx context.Context) (keys.Authorizer, error) {
	adminAcc := flow.HexToAddress(s.cfg.AdminAddress)

	index, err := s.store.ProposalKey()
	if err != nil {
		return keys.Authorizer{}, err
	}

	acc, err := s.fc.GetAccount(ctx, adminAcc)
	if err != nil {
		return keys.Authorizer{}, err
	}

	sig, err := signerForKey(ctx, adminAcc, s.adminAccountKey)
	if err != nil {
		return keys.Authorizer{}, err
	}

	return keys.Authorizer{
		Address: adminAcc,
		Key:     acc.Keys[index],
		Signer:  sig,
	}, nil
}

func signerForKey(ctx context.Context, address flow.Address, k keys.Private) (crypto.Signer, error) {
	var (
		sig crypto.Signer
		err error
	)

	switch k.Type {
	default:
		return nil, fmt.Errorf("key.Type not recognised: %s", k.Type)
	case keys.AccountKeyTypeLocal:
		sig, err = local.Signer(k)
		if err != nil {
			return nil, err
		}
	case keys.AccountKeyTypeGoogleKMS:
		sig, err = google.Signer(ctx, address, k)
		if err != nil {
			return nil, err
		}
	}

	return sig, nil
}
