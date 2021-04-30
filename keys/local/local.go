package local

import (
	"crypto/rand"

	"github.com/eqlabs/flow-nft-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Generate(
	signAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
	keyIndex, weight int,
) (result keys.Wrapped, err error) {
	seed := make([]byte, crypto.MinSeedLength)
	_, err = rand.Read(seed)
	if err != nil {
		return
	}

	privateKey, err := crypto.GeneratePrivateKey(signAlgo, seed)
	if err != nil {
		return
	}

	flowKey := flow.NewAccountKey().
		FromPrivateKey(privateKey).
		SetHashAlgo(hashAlgo).
		SetWeight(weight)

	flowKey.Index = keyIndex

	key := keys.Key{
		Index: keyIndex,
		Type:  keys.ACCOUNT_KEY_TYPE_LOCAL,
		Value: privateKey.String(),
	}

	result.AccountKey = key
	result.FlowKey = flowKey

	return
}

func Signer(
	signAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
	key keys.Key,
) (result crypto.Signer, err error) {
	pk, err := crypto.DecodePrivateKeyHex(signAlgo, key.Value)
	if err != nil {
		return
	}
	result = crypto.NewInMemorySigner(pk, hashAlgo)
	return
}
