// Package local provides functions for local key and signer generation.
package local

import (
	"crypto/rand"
	"strings"

	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Generate(
	keyIndex, weight int,
	signAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (keys.Wrapped, error) {
	s := make([]byte, crypto.MinSeedLength)

	_, err := rand.Read(s)
	if err != nil {
		return keys.Wrapped{}, err
	}

	pk, err := crypto.GeneratePrivateKey(signAlgo, s)
	if err != nil {
		return keys.Wrapped{}, err
	}

	f := flow.NewAccountKey().
		FromPrivateKey(pk).
		SetHashAlgo(hashAlgo).
		SetWeight(weight)

	f.Index = keyIndex

	p := keys.Private{
		Index:    keyIndex,
		Type:     keys.ACCOUNT_KEY_TYPE_LOCAL,
		Value:    strings.TrimPrefix(pk.String(), "0x"),
		SignAlgo: signAlgo,
		HashAlgo: hashAlgo,
	}

	return keys.Wrapped{
		AccountKey: f,
		PrivateKey: p,
	}, nil
}

func Signer(key keys.Private) (crypto.Signer, error) {
	p, err := crypto.DecodePrivateKeyHex(key.SignAlgo, key.Value)
	if err != nil {
		return crypto.InMemorySigner{}, err
	}
	return crypto.NewInMemorySigner(p, key.HashAlgo), nil
}
