package test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/onflow/flow-go-sdk"
	access "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	seed_length = 64
)

var (
	max_tx_wait = 10 * time.Second
)

func NewFlowClient(t *testing.T, cfg *configs.Config) flow_helpers.FlowClient {
	fc, err := access.NewClient(cfg.AccessAPIHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	if err := fc.Ping(context.Background()); err != nil {
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

func NewFlowAccount(t *testing.T, fc flow_helpers.FlowClient, creatorAddress flow.Address, creatorKey *flow.AccountKey, creatorSigner crypto.Signer) *flow.Account {
	seed := make([]byte, seed_length)
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
	tx, err := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, creatorAddress)
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

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
