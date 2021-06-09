// Package account provides functions for account management on Flow blockhain.
package accounts

import (
	"context"
	"fmt"
	"time"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/templates/template_strings"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	flow_templates "github.com/onflow/flow-go-sdk/templates"
	"gorm.io/gorm"
)

// Account struct represents a storable account.
type Account struct {
	Address   string          `json:"address" gorm:"primaryKey"`
	Keys      []keys.Storable `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Tokens    []AccountToken  `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time       `json:"createdAt" `
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index"`
}

type AccountToken struct {
	ID             int            `json:"-" gorm:"primaryKey"`
	AccountAddress string         `json:"-" gorm:"index"`
	Name           string         `json:"name"`
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// TODO: Add AccountTokens to admin account on startup (FlowToken for now)

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
	accountKey, newPrivateKey, err := km.GenerateDefault(ctx)
	if err != nil {
		return
	}

	tx := flow_templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, auth.Address)
	b := templates.NewBuilderFromTx(tx)

	t, err := transactions.New(id, b, transactions.General, auth, auth, nil)
	if err != nil {
		return
	}

	err = t.SendAndWait(ctx, fc)
	if err != nil {
		return
	}

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

// AddContract is mainly used for testing purposes
func AddContract(
	ctx context.Context,
	fc *client.Client,
	km keys.Manager,
	accountAddress string,
	contract flow_templates.Contract) (*transactions.Transaction, error) {

	// Get admin account authorizer
	adminAuth, err := km.AdminAuthorizer(ctx)
	if err != nil {
		return nil, err
	}

	flowAddr := flow.HexToAddress(accountAddress)

	// Get user account authorizer
	userAuth, err := km.UserAuthorizer(ctx, flowAddr)
	if err != nil {
		return nil, err
	}

	// Get latest blocks id as reference id
	id, err := flow_helpers.LatestBlockId(ctx, fc)
	if err != nil {
		return nil, err
	}

	raw := templates.Raw{
		Code: template_strings.AddAccountContractWithAdmin,
		Arguments: []templates.Argument{
			cadence.NewString(contract.Name),
			cadence.NewString(contract.SourceHex()),
		},
	}

	b, err := templates.NewBuilderFromRaw(raw)
	if err != nil {
		return nil, err
	}

	b.Tx.AddAuthorizer(adminAuth.Address)

	t, err := transactions.New(id, b, transactions.General, userAuth, adminAuth, nil)
	if err != nil {
		return nil, err
	}

	err = t.SendAndWait(ctx, fc)
	if err != nil {
		return nil, err
	}

	return t, nil
}
