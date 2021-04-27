package tokens

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/flow_helpers"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/keys"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

type NftInfo struct {
	BaseAddress string
	Address     string
	Name        string
}

func SetupNFT(ctx context.Context, fc *client.Client, ks keys.KeyStore, address string, n NftInfo) (*flow.TransactionResult, error) {
	serviceAuth, err := ks.ServiceAuthorizer(ctx, fc)
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	accountAuth, err := ks.AccountAuthorizer(ctx, fc, address)
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	txTemplate, err := ioutil.ReadFile("../../cadence/transactions/setup_account.cdc")
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	replacer := strings.NewReplacer(
		"<BaseNFTAddress>", "0x"+n.BaseAddress,
		"<NFTAddress>", "0x"+n.Address,
		"<NFTName>", n.Name)

	txStr := replacer.Replace(string(txTemplate))

	referenceBlockID, err := flow_helpers.GetLatestBlockId(ctx, fc)
	if err != nil {
		return &flow.TransactionResult{}, err
	}
	tx := flow.NewTransaction().
		SetScript([]byte(txStr)).
		SetGasLimit(100).
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(accountAuth.Address, accountAuth.Key.Index, accountAuth.Key.SequenceNumber).
		SetPayer(serviceAuth.Address).
		AddAuthorizer(accountAuth.Address)

	// Proposer signs the payload first
	err = tx.SignPayload(accountAuth.Address, accountAuth.Key.Index, accountAuth.Signer)
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	// Sign the transaction with the service account
	err = tx.SignEnvelope(serviceAuth.Address, serviceAuth.Key.Index, serviceAuth.Signer)
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	// Send the transaction to the network
	err = fc.SendTransaction(ctx, *tx)
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	result, err := flow_helpers.WaitForSeal(ctx, fc, tx.ID())
	if err != nil {
		return &flow.TransactionResult{}, err
	}

	return result, nil
}
