package tokens

import (
	"context"

	"github.com/eqlabs/flow-wallet-service/flow_helpers"
	"github.com/eqlabs/flow-wallet-service/keys"
	"github.com/eqlabs/flow-wallet-service/transactions"
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

	code, err := ParseTransferFlowToken(flow.Emulator)
	if err != nil {
		return flow.EmptyID, err
	}

	aa := make([]transactions.Argument, 2)

	c_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return flow.EmptyID, err
	}

	aa[0] = c_amount
	aa[1] = cadence.NewAddress(recipientAddress)

	id, err := flow_helpers.LatestBlockId(context.Background(), fc)
	if err != nil {
		return flow.EmptyID, err
	}

	auth, err := km.UserAuthorizer(ctx, senderAddress)
	if err != nil {
		return flow.EmptyID, err
	}

	t, err := transactions.New(id, code, aa, auth, auth, []keys.Authorizer{auth})
	if err != nil {
		return flow.EmptyID, err
	}

	t.Send(ctx, fc)

	return flow.HexToID(t.TransactionId), nil
}
