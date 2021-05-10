package tokens

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

var EmulatorFlowToken Token = Token{
	"FlowToken",
	"0xee82856bf20e2aa6",
	"0x0ae53cb6e3f42a79",
}

func TransferFlow(
	km keys.Manager,
	fc *client.Client,
	recipientAddress flow.Address,
	senderAddress flow.Address,
	amount string) (flow.Identifier, error) {

	tmplPath := filepath.Join(TemplatePath(), "transactions", "transfer_flow.cdc")
	txTemplate, err := ioutil.ReadFile(tmplPath)
	if err != nil {
		return flow.EmptyID, err
	}

	replacer := strings.NewReplacer(
		"<BaseTokenAddress>", EmulatorFlowToken.BaseAddress,
		"<TokenAddress>", EmulatorFlowToken.Address,
	)

	txStr := replacer.Replace(string(txTemplate))

	senderAuthorizer, err := km.UserAuthorizer(senderAddress.Hex())
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
