package tokens

import (
	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func parseFtWithdrawal(
	tokenName,
	recipientAddress,
	amount string,
	chainId flow.ChainID,
	contractAddresses ...string) (templates.Raw, error) {
	// Check if the input is a valid address
	err := flow_helpers.ValidateAddress(recipientAddress, chainId)
	if err != nil {
		return templates.Raw{}, err
	}

	c := templates.GenericFungibleTransferCode(
		tokenName,
		chainId,
		contractAddresses...,
	)

	aa := make([]templates.Argument, 2)

	_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return templates.Raw{}, err
	}

	aa[0] = _amount
	aa[1] = cadence.NewAddress(flow.HexToAddress(recipientAddress))

	return templates.Raw{Code: c, Arguments: aa}, nil
}

func parseFtSetup(
	tokenName string,
	chainId flow.ChainID,
	contractAddresses ...string) (templates.Raw, error) {
	c := templates.GenericFungibleSetupCode(
		tokenName,
		chainId,
		contractAddresses...,
	)
	return templates.Raw{Code: c}, nil
}
