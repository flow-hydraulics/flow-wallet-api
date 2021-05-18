// Package account provides functions for account management on Flow blockhain.
package accounts

import (
	"context"
	"fmt"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
	"gorm.io/gorm"
)

// Account struct represents a storable account.
type Account struct {
	Address   string          `json:"address" gorm:"primaryKey"`
	Keys      []keys.Storable `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt time.Time       `json:"createdAt" `
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index"`
}

// New creates a new account on the Flow blockchain.
// It uses the provided admin account to pay for the creation.
// It generates a new privatekey and returns it (local key)
// or a reference to it (Google KMS resource id).
func New(
	ctx context.Context,
	fc *client.Client,
	km keys.Manager,
) (
	newAccount Account,
	newPrivateKey keys.Private,
	err error,
) {
	// Get admin account authorizer
	adminAuth, err := km.AdminAuthorizer(ctx)
	if err != nil {
		return
	}

	// Get latest blocks id as reference id
	referenceBlockID, err := flow_helpers.LatestBlockId(ctx, fc)
	if err != nil {
		return
	}

	// Generate a new key pair
	wrapped, err := km.GenerateDefault(ctx)
	if err != nil {
		return
	}

	// Destruct the wrapped key
	publicKey := wrapped.AccountKey
	newPrivateKey = wrapped.PrivateKey

	// Setup a transaction to create an account
	tx := templates.
		CreateAccount([]*flow.AccountKey{publicKey}, nil, adminAuth.Address).
		SetProposalKey(adminAuth.Address, adminAuth.Key.Index, adminAuth.Key.SequenceNumber).
		SetReferenceBlockID(referenceBlockID).
		SetPayer(adminAuth.Address)

	// Sign the transaction with the service account
	err = tx.SignEnvelope(adminAuth.Address, adminAuth.Key.Index, adminAuth.Signer)
	if err != nil {
		return
	}

	// Send the transaction to the network
	err = fc.SendTransaction(ctx, *tx)
	if err != nil {
		return
	}

	// Wait for the transaction to be sealed
	result, err := flow_helpers.WaitForSeal(ctx, fc, tx.ID())
	if err != nil {
		return
	}

	// Grab the new address from transaction events
	var newAddress string
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address().Hex()
			break
		}
	}

	// Check that we actually got a new address
	if newAddress == (flow.Address{}.Hex()) {
		err = fmt.Errorf("something went wrong when waiting for address")
		return
	}

	newAccount.Address = newAddress

	return
}
