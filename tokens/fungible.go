package tokens

import (
	"context"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

func TransferFlow(
	ctx context.Context,
	km keys.Manager,
	fc *client.Client,
	recipientAddress flow.Address,
	senderAddress flow.Address,
	amount string) (flow.Identifier, error) {

	txStr, err := ParseTransferFlowToken(flow.Emulator)
	if err != nil {
		return flow.EmptyID, err
	}

	senderAuthorizer, err := km.UserAuthorizer(ctx, senderAddress.Hex())
	if err != nil {
		return flow.EmptyID, err
	}

	tx, err := makeTransferTokensTx(recipientAddress, senderAddress, *senderAuthorizer.Key, amount, txStr)
	if err != nil {
		return flow.EmptyID, err
	}

	referenceBlockID, err := flow_helpers.LatestBlockId(context.Background(), fc)
	if err != nil {
		return flow.EmptyID, err
	}

	tx.SetReferenceBlockID(referenceBlockID)

	err = tx.SignEnvelope(senderAddress, senderAuthorizer.Key.Index, senderAuthorizer.Signer)
	if err != nil {
		return flow.EmptyID, err
	}

	err = fc.SendTransaction(context.Background(), *tx)
	if err != nil {
		return flow.EmptyID, err
	}

	return tx.ID(), nil
}

func makeTransferTokensTx(
	recipientAddress flow.Address,
	senderAddress flow.Address,
	senderAccountKey flow.AccountKey,
	amount string,
	transferTemplate string) (*flow.Transaction, error) {

	tx := flow.NewTransaction().
		SetScript([]byte(transferTemplate)).
		SetGasLimit(100).
		SetPayer(senderAddress).
		AddAuthorizer(senderAddress).
		SetProposalKey(senderAddress, senderAccountKey.Index, senderAccountKey.SequenceNumber)

	amountFixed, err := cadence.NewUFix64(amount)
	if err != nil {
		return &flow.Transaction{}, err
	}

	recipient := cadence.NewAddress(recipientAddress)

	err = tx.AddArgument(amountFixed)
	if err != nil {
		return &flow.Transaction{}, err
	}

	err = tx.AddArgument(recipient)
	if err != nil {
		return &flow.Transaction{}, err
	}

	return tx, nil
}
