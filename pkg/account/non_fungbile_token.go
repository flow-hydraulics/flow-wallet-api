package account

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

func (a Account) SetupNFT(ctx context.Context, c *client.Client, sAcct Account, n NFT) *flow.TransactionResult {
	txTemplate, err := ioutil.ReadFile("../../cadence/transactions/setup_account.cdc")
	handle(err)

	replacer := strings.NewReplacer(
		"<BaseNFTAddress>", "0x"+n.BaseAddress,
		"<NFTAddress>", "0x"+n.Address,
		"<NFTName>", n.Name)

	txStr := replacer.Replace(string(txTemplate))

	proposerAcctAddr, proposerAcctKey, proposerSigner := authorize(ctx, c, a)
	serviceAcctAddr, serviceAcctKey, serviceSigner := authorize(ctx, c, sAcct)

	referenceBlockID := getReferenceBlockId(ctx, c)

	tx := flow.NewTransaction().
		SetScript([]byte(txStr)).
		SetGasLimit(100).
		SetReferenceBlockID(referenceBlockID).
		SetProposalKey(
			proposerAcctAddr,
			proposerAcctKey.Index,
			proposerAcctKey.SequenceNumber).
		SetPayer(serviceAcctAddr).
		AddAuthorizer(proposerAcctAddr)

	// Proposer signs the payload first
	err = tx.SignPayload(proposerAcctAddr, proposerAcctKey.Index, proposerSigner)
	handle(err)

	// Sign the transaction with the service account
	err = tx.SignEnvelope(serviceAcctAddr, serviceAcctKey.Index, serviceSigner)
	handle(err)

	// Send the transaction to the network
	err = c.SendTransaction(ctx, *tx)
	handle(err)

	result := waitForSeal(ctx, c, tx.ID())

	return result
}
