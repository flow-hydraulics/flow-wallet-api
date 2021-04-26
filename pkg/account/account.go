package account

import (
	"context"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/flow_helpers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
)

type AccountService interface {
	Create(context.Context, *client.Client, store.KeyStore)
}

func Create(ctx context.Context, fc *client.Client, ks store.KeyStore) (flow.Address, error) {
	serviceAuth, err := ks.ServiceAuthorizer(ctx, fc)
	if err != nil {
		return flow.Address{}, err
	}

	referenceBlockID, err := flow_helpers.GetLatestBlockId(ctx, fc)
	if err != nil {
		return flow.Address{}, err
	}

	// Generate a new key pair
	newKey, err := ks.Generate(ctx, 0, flow.AccountKeyWeightThreshold)
	if err != nil {
		return flow.Address{}, err
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
		return flow.Address{}, err
	}

	// Send the transaction to the network
	err = fc.SendTransaction(ctx, *tx)
	if err != nil {
		// TODO: check what needs to be reverted
		return flow.Address{}, err
	}

	// Wait for the transaction to be sealed
	result, err := flow_helpers.WaitForSeal(ctx, fc, tx.ID())
	if err != nil {
		// TODO: check what needs to be reverted
		return flow.Address{}, err
	}

	// Grab the new address from transaction events
	var newAddress flow.Address
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address()
		}
	}

	// Store the new key
	newKey.AccountKey.AccountAddress = newAddress
	if err = ks.Save(newKey.AccountKey); err != nil {
		// TODO: check what needs to be reverted
		return flow.Address{}, err
	}

	return newAddress, nil
}
