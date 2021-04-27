package account

import (
	"context"
	"fmt"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/flow_helpers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
)

func New(ctx context.Context, fc *client.Client, ks store.KeyStore) (store.Account, store.AccountKey, error) {
	serviceAuth, err := ks.ServiceAuthorizer(ctx, fc)
	if err != nil {
		return store.Account{}, store.AccountKey{}, err
	}

	referenceBlockID, err := flow_helpers.GetLatestBlockId(ctx, fc)
	if err != nil {
		return store.Account{}, store.AccountKey{}, err
	}

	// Generate a new key pair
	newKey, err := ks.Generate(ctx, 0, flow.AccountKeyWeightThreshold)
	if err != nil {
		return store.Account{}, store.AccountKey{}, err
	}

	// Setup a transaction to create an account
	tx := templates.CreateAccount([]*flow.AccountKey{newKey.FlowKey}, nil, serviceAuth.Address)
	tx.SetProposalKey(serviceAuth.Address, serviceAuth.Key.Index, serviceAuth.Key.SequenceNumber)
	tx.SetReferenceBlockID(referenceBlockID)
	tx.SetPayer(serviceAuth.Address)

	// Sign the transaction with the service account
	err = tx.SignEnvelope(serviceAuth.Address, serviceAuth.Key.Index, serviceAuth.Signer)
	if err != nil {
		// TODO: check what needs to be reverted
		return store.Account{}, store.AccountKey{}, err
	}

	// Send the transaction to the network
	err = fc.SendTransaction(ctx, *tx)
	if err != nil {
		// TODO: check what needs to be reverted
		return store.Account{}, store.AccountKey{}, err
	}

	// Wait for the transaction to be sealed
	result, err := flow_helpers.WaitForSeal(ctx, fc, tx.ID())
	if err != nil {
		// TODO: check what needs to be reverted
		return store.Account{}, store.AccountKey{}, err
	}

	// Grab the new address from transaction events
	var newAddress flow.Address
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address()
		}
	}

	if (flow.Address{}) == newAddress {
		// TODO: check what needs to be reverted
		return store.Account{}, store.AccountKey{}, fmt.Errorf("something went wrong when waiting for address")
	}

	newKey.AccountKey.AccountAddress = newAddress

	return store.Account{Address: newAddress}, newKey.AccountKey, nil
}
