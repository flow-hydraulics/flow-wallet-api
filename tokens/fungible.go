package tokens

import (
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func parseFtWithdrawal(
	tokenName,
	recipientAddress,
	amount string,
	chainId flow.ChainID,
	contractAddresses ...string) (string, []transactions.Argument, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(recipientAddress, chainId)
	if err != nil {
		return "", transactions.EmptyArguments, err
	}

	c := templates.ParseGenericFungibleTransfer(
		tokenName,
		chainId,
		contractAddresses...,
	)

	aa := make([]transactions.Argument, 2)

	_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return "", transactions.EmptyArguments, err
	}

	aa[0] = _amount
	aa[1] = cadence.NewAddress(flow.HexToAddress(recipientAddress))

	return c, aa, nil
}

func parseFtSetup(
	tokenName string,
	chainId flow.ChainID,
	contractAddresses ...string) (string, []transactions.Argument, error) {
	c := templates.ParseGenericFungibleSetup(
		tokenName,
		chainId,
		contractAddresses...,
	)
	return c, []transactions.Argument{}, nil
}
