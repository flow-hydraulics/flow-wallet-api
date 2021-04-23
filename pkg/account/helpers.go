package account

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

type NFT struct {
	BaseAddress string
	Address     string
	Name        string
}

type FlowJsonConfig struct {
	Accounts map[string]FlowJsonConfigAccount `json:"accounts"`
}

type FlowJsonConfigAccount struct {
	Address string `json:"address"`
	Keys    string `json:"keys"`
}

func handle(err error) {
	if err != nil {
		fmt.Println("err:", err.Error())
		panic(err)
	}
}

func authorize(ctx context.Context, c *client.Client, authAcct Account) (flow.Address, *flow.AccountKey, crypto.Signer) {
	privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, authAcct.PrivateKey)
	handle(err)

	addr := flow.HexToAddress(authAcct.Address)
	acc, err := c.GetAccount(ctx, addr)
	handle(err)

	accountKey := acc.Keys[0]
	signer := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)

	return addr, accountKey, signer
}

func randomPrivateKey() crypto.PrivateKey {
	seed := make([]byte, crypto.MinSeedLength)
	_, err := rand.Read(seed)
	handle(err)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	handle(err)

	return privateKey
}

func getReferenceBlockId(ctx context.Context, c *client.Client) flow.Identifier {
	block, err := c.GetLatestBlock(ctx, true)
	handle(err)

	return block.ID
}

func waitForSeal(ctx context.Context, c *client.Client, id flow.Identifier) *flow.TransactionResult {
	result, err := c.GetTransactionResult(ctx, id)
	handle(err)

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, id)
		handle(err)
	}

	return result
}

func NewFromFlowFile(configPath string, acctName string) (*Account, error) {
	jsonFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var conf FlowJsonConfig
	err = json.Unmarshal(jsonFile, &conf)
	if err != nil {
		return nil, err
	}

	acct := conf.Accounts[acctName]

	return &Account{
		Address:    acct.Address,
		PrivateKey: acct.Keys,
	}, nil
}
