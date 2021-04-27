package account

import (
	"context"
	"fmt"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/data"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/flow_helpers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
)

func Create(ctx context.Context, fc *client.Client, ks keys.Store) (data.Account, data.AccountKey, error) {
	serviceAuth, err := ks.ServiceAuthorizer(ctx, fc)
	if err != nil {
		return data.Account{}, data.AccountKey{}, err
	}

	referenceBlockID, err := flow_helpers.GetLatestBlockId(ctx, fc)
	if err != nil {
		return data.Account{}, data.AccountKey{}, err
	}

	// Generate a new key pair
	newKey, err := ks.Generate(ctx, 0, flow.AccountKeyWeightThreshold)
	if err != nil {
		return data.Account{}, data.AccountKey{}, err
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
		return data.Account{}, data.AccountKey{}, err
	}

	// Send the transaction to the network
	err = fc.SendTransaction(ctx, *tx)
	if err != nil {
		// TODO: check what needs to be reverted
		return data.Account{}, data.AccountKey{}, err
	}

	// Wait for the transaction to be sealed
	result, err := flow_helpers.WaitForSeal(ctx, fc, tx.ID())
	if err != nil {
		// TODO: check what needs to be reverted
		return data.Account{}, data.AccountKey{}, err
	}

	// Grab the new address from transaction events
	var newAddress string
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address().Hex()
		}
	}

	if (flow.Address{}.Hex()) == newAddress {
		// TODO: check what needs to be reverted
		return data.Account{}, data.AccountKey{}, fmt.Errorf("something went wrong when waiting for address")
	}

	newKey.AccountKey.AccountAddress = newAddress

	return data.Account{Address: newAddress}, newKey.AccountKey, nil
}
