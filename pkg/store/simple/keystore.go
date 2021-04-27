package simple

import (
	"context"
	"fmt"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

type KeyStore struct {
	db            store.DataStore
	serviceAcct   store.AccountKey
	encryptionKey string
	signAlgo      crypto.SignatureAlgorithm
	hashAlgo      crypto.HashAlgorithm
}

func NewKeyStore(
	db store.DataStore,
	serviceAcct store.AccountKey,
	encryptionKey string,
) (*KeyStore, error) {
	return &KeyStore{
		db,
		serviceAcct,
		encryptionKey,
		crypto.ECDSA_P256, // TODO: config
		crypto.SHA3_256,   // TODO: config
	}, nil
}

func (s *KeyStore) Generate(ctx context.Context, keyIndex int, weight int) (store.NewKeyWrapper, error) {
	// TODO: get account key type
	panic("not implemented") // TODO: implement
}

func (s *KeyStore) Save(store.AccountKey) error {
	panic("not implemented") // TODO: implement
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
		// TODO: decrypt accountKey.Value
	}

	flowAcc, err := fc.GetAccount(ctx, addr)
	if err != nil {
		return authorizer, err
	}

	authorizer.Key = flowAcc.Keys[accountKey.Index]

	// TODO: Decide whether we want to allow this kind of flexibility
	// or should we just panic if `AccountKey.Type` != configured type
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
			fmt.Errorf("AccountKey.Type not recognised: %s", accountKey.Type)
	}

	return authorizer, nil
}
