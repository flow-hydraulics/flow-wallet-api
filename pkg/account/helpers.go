package account

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/config"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Handle(err error) {
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}
}

func ServiceAccount(flowClient *client.Client, confAcc config.FlowConfigAccount) (flow.Address, *flow.AccountKey, crypto.Signer) {
	sigAlgo := crypto.StringToSignatureAlgorithm(confAcc.SigAlgo)
	privateKey, err := crypto.DecodePrivateKeyHex(sigAlgo, confAcc.Keys)
	Handle(err)

	addr := flow.HexToAddress(confAcc.Address)
	acc, err := flowClient.GetAccount(context.Background(), addr)
	Handle(err)

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)
	return addr, accountKey, signer
}

func RandomPrivateKey() crypto.PrivateKey {
	seed := make([]byte, crypto.MinSeedLength)
	_, err := rand.Read(seed)
	Handle(err)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	Handle(err)

	return privateKey
}

func GetReferenceBlockId(flowClient *client.Client) flow.Identifier {
	block, err := flowClient.GetLatestBlock(context.Background(), true)
	Handle(err)

	return block.ID
}

func WaitForSeal(ctx context.Context, c *client.Client, id flow.Identifier) *flow.TransactionResult {
	result, err := c.GetTransactionResult(ctx, id)
	Handle(err)

	fmt.Printf("Waiting for transaction %s to be sealed...\n", id)

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = c.GetTransactionResult(ctx, id)
		Handle(err)
	}

	fmt.Println()
	fmt.Printf("Transaction %s sealed\n", id)

	return result
}
