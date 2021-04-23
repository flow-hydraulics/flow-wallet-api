package account

import (
	"context"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

type Account struct {
	Address    string
	PrivateKey string
}

func Create(ctx context.Context, c *client.Client, sAcct Account) Account {
	serviceAcctAddr, serviceAcctKey, serviceSigner := authorize(ctx, c, sAcct)

	newPrivateKey := randomPrivateKey()
	newAcctKey := flow.NewAccountKey().
		FromPrivateKey(newPrivateKey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flow.AccountKeyWeightThreshold)

	referenceBlockID := getReferenceBlockId(ctx, c)
	tx := templates.CreateAccount([]*flow.AccountKey{newAcctKey}, nil, serviceAcctAddr)
	tx.SetProposalKey(
		serviceAcctAddr,
		serviceAcctKey.Index,
		serviceAcctKey.SequenceNumber,
	)
	tx.SetReferenceBlockID(referenceBlockID)
	tx.SetPayer(serviceAcctAddr)

	// Sign the transaction with the service account
	err := tx.SignEnvelope(serviceAcctAddr, serviceAcctKey.Index, serviceSigner)
	handle(err)

	// Send the transaction to the network
	err = c.SendTransaction(ctx, *tx)
	handle(err)

	result := waitForSeal(ctx, c, tx.ID())

	var newAddress flow.Address

	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address()
		}
	}

	return Account{
		Address:    newAddress.Hex(),
		PrivateKey: strings.TrimPrefix(newPrivateKey.String(), "0x"),
	}
}
