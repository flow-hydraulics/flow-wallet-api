// Package local provides functions for local key and signer generation.
package local

import (
	"context"
	"crypto/rand"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Generate(
	keyIndex, weight int,
	signAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (*flow.AccountKey, *keys.Private, error) {
	s := make([]byte, crypto.MinSeedLength)

	_, err := rand.Read(s)
	if err != nil {
		return nil, nil, err
	}

	pk, err := crypto.GeneratePrivateKey(signAlgo, s)
	if err != nil {
		return nil, nil, err
	}

	f := flow.NewAccountKey().
		FromPrivateKey(pk).
		SetHashAlgo(hashAlgo).
		SetWeight(weight)

	f.Index = keyIndex

	p := &keys.Private{
		Index:    keyIndex,
		Type:     keys.AccountKeyTypeLocal,
		Value:    strings.TrimPrefix(pk.String(), "0x"),
		SignAlgo: signAlgo,
		HashAlgo: hashAlgo,
	}

	return f, p, nil
}

func Signer(ctx context.Context, key keys.Private) (crypto.Signer, error) {
	p, err := crypto.DecodePrivateKeyHex(key.SignAlgo, key.Value)
	if err != nil {
		return crypto.InMemorySigner{}, err
	}
	return crypto.NewInMemorySigner(p, key.HashAlgo)
}
