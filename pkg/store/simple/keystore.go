package simple

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

type KeyStore struct {
	db                store.DataStore
	serviceAcct       store.AccountKey
	defaultKeyManager string
	encryptionKey     string
	signAlgo          crypto.SignatureAlgorithm
	hashAlgo          crypto.HashAlgorithm
}

func NewKeyStore(
	db store.DataStore,
	serviceAcct store.AccountKey,
	defaultKeyManager string,
	encryptionKey string,
) (*KeyStore, error) {
	return &KeyStore{
		db,
		serviceAcct,
		defaultKeyManager,
		encryptionKey,
		crypto.ECDSA_P256, // TODO: config
		crypto.SHA3_256,   // TODO: config
	}, nil
}

func (s *KeyStore) Generate(ctx context.Context, keyIndex int, weight int) (store.NewKeyWrapper, error) {
	switch s.defaultKeyManager {
	case store.ACCOUNT_KEY_TYPE_LOCAL:
		seed := make([]byte, crypto.MinSeedLength)
		_, err := rand.Read(seed)
		if err != nil {
			return store.NewKeyWrapper{}, err
		}
		privateKey, err := crypto.GeneratePrivateKey(s.signAlgo, seed)
		if err != nil {
			return store.NewKeyWrapper{}, err
		}

		flowKey := flow.NewAccountKey().
			FromPrivateKey(privateKey).
			SetHashAlgo(s.hashAlgo).
			SetWeight(weight)
		flowKey.Index = keyIndex

		accountKey := store.AccountKey{
			Index: keyIndex,
			Type:  store.ACCOUNT_KEY_TYPE_LOCAL,
			Value: privateKey.String(),
		}
		return store.NewKeyWrapper{FlowKey: flowKey, AccountKey: accountKey}, nil
	default:
		// TODO: google_kms
		return store.NewKeyWrapper{}, fmt.Errorf("keyStore.Generate() not implmented for %s", s.defaultKeyManager)
	}
}

func (s *KeyStore) Save(key store.AccountKey) error {
	switch key.Type {
	case store.ACCOUNT_KEY_TYPE_LOCAL:
		// TODO: encrypt key.Value
		if s.encryptionKey != "" {
			panic("key encryption not implemented")
		}
		err := s.db.InsertAccountKey(key)
		return err
	default:
		// TODO: google_kms
		return fmt.Errorf("keyStore.Save() not implmented for %s", s.defaultKeyManager)
	}
}

func (s *KeyStore) Delete(addr flow.Address, keyIndex int) error {
	panic("not implemented") // TODO: implement
}

func (s *KeyStore) ServiceAuthorizer(ctx context.Context, fc *client.Client) (store.Authorizer, error) {
	return s.MakeAuthorizer(ctx, fc, s.serviceAcct.AccountAddress)
}

func (s *KeyStore) AccountAuthorizer(ctx context.Context, fc *client.Client, addr flow.Address) (store.Authorizer, error) {
	return s.MakeAuthorizer(ctx, fc, addr)
}

func (s *KeyStore) MakeAuthorizer(ctx context.Context, fc *client.Client, addr flow.Address) (store.Authorizer, error) {
	var (
		accountKey store.AccountKey
		authorizer store.Authorizer = store.Authorizer{}
	)

	authorizer.Address = addr

	if addr == s.serviceAcct.AccountAddress {
		accountKey = s.serviceAcct
	} else {
		ak, err := s.db.AccountKey(addr)
		if err != nil {
			return authorizer, err
		}
		accountKey = ak
		if s.encryptionKey != "" {
			// TODO: decrypt accountKey.Value
			panic("key decryption not implemented")
		}
	}

	flowAcc, err := fc.GetAccount(ctx, addr)
	if err != nil {
		return authorizer, err
	}

	authorizer.Key = flowAcc.Keys[accountKey.Index]

	// TODO: Decide whether we want to allow this kind of flexibility
	// or should we just panic if `accountKey.Type` != `s.defaultKeyManager`
	switch accountKey.Type {
	case store.ACCOUNT_KEY_TYPE_LOCAL:
		pk, err := crypto.DecodePrivateKeyHex(s.signAlgo, accountKey.Value)
		if err != nil {
			return authorizer, err
		}
		authorizer.Signer = crypto.NewInMemorySigner(pk, s.hashAlgo)
	case store.ACCOUNT_KEY_TYPE_GOOGLE_KMS:
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
			addr,
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
