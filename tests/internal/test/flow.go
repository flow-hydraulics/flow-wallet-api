package test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
)

const (
	seed_length = 64
)

var (
	max_tx_wait = 10 * time.Second
)

func NewFlowClient(t *testing.T, cfg *configs.Config) *client.Client {
	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	close := func() {
		err := fc.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(close)

	return fc
}

func NewFlowAccount(t *testing.T, fc *client.Client, creatorAddress flow.Address, creatorKey *flow.AccountKey, creatorSigner crypto.Signer) *flow.Account {
	seed := make([]byte, seed_length, seed_length)
	readRandom(t, seed)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := privateKey.PublicKey()

	accountKey := flow.NewAccountKey().
		SetPublicKey(publicKey).                  // The signature algorithm is inferred from the public key
		SetHashAlgo(crypto.SHA3_256).             // This key will require SHA3 hashes
		SetWeight(flow.AccountKeyWeightThreshold) // Give this key full signing weight

	// Use the templates package to create a new account creation transaction
	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, creatorAddress)

	// Set the transaction payer and proposal key
	tx.SetPayer(creatorAddress)
	tx.SetProposalKey(
		creatorAddress,
		creatorKey.Index,
		creatorKey.SequenceNumber,
	)

	// Get the latest sealed block to use as a reference block
	latestBlock, err := fc.GetLatestBlockHeader(context.Background(), true)
	if err != nil {
		panic("failed to fetch latest block")
	}

	tx.SetReferenceBlockID(latestBlock.ID)

	// Sign and submit the transaction

	err = tx.SignEnvelope(creatorAddress, creatorKey.Index, creatorSigner)
	if err != nil {
		panic("failed to sign transaction envelope")
	}

	err = fc.SendTransaction(context.Background(), *tx)
	if err != nil {
		panic("failed to send transaction to network")
	}

	result, err := flow_helpers.WaitForSeal(context.Background(), fc, tx.ID(), max_tx_wait)

	var newAddress flow.Address
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			newAddress = flow.AccountCreatedEvent(event).Address()
			break
		}
	}

	a, err := fc.GetAccount(context.Background(), newAddress)
	if err != nil {
		t.Fatal(err)
	}

	return a
}

func readRandom(t *testing.T, buf []byte) {
	var bytesRead int

	for bytesRead < len(buf) {
		n, err := rand.Read(buf[bytesRead:])
		if err != nil {
			t.Fatal(err)
		}
		bytesRead += n
	}
}
