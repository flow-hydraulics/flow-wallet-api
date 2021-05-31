// Package account provides functions for account management on Flow blockhain.
package accounts

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
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
	auth, err := km.AdminAuthorizer(ctx)
	if err != nil {
		return
	}

	// Get latest blocks id as reference id
	id, err := flow_helpers.LatestBlockId(ctx, fc)
	if err != nil {
		return
	}

	// Generate a new key pair
	wrapped, err := km.GenerateDefault(ctx)
	if err != nil {
		return
	}

	// Destruct the wrapped key
	accountKey := wrapped.AccountKey
	newPrivateKey = wrapped.PrivateKey

	aa := []transactions.Argument{
		cadence.NewArray([]cadence.Value{
			cadence.NewString(hex.EncodeToString(accountKey.Encode())),
		})}

	t, err := transactions.New(
		id,
		templates.CreateAccount,
		aa,
		transactions.General,
		auth, auth, []keys.Authorizer{auth},
	)

	t.SendAndWait(ctx, fc)

	// Grab the new address from transaction events
	var newAddress flow.Address
	for _, event := range t.Result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address()
			break
		}
	}

	// Check that we actually got a new address
	if newAddress == flow.EmptyAddress {
		err = fmt.Errorf("something went wrong when waiting for address")
		return
	}

	newAccount.Address = flow_helpers.FormatAddress(newAddress)

	return
}
