package tokens

import (
	"context"

	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/eqlabs/flow-wallet-service/transactions"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func TransferFlow(
	ctx context.Context,
	s *transactions.Service,
	recipientAddress,
	senderAddress,
	amount string) (*transactions.Transaction, error) {

	c := templates.ParseCode(templates.TransferFlow, flow.Emulator)

	aa := make([]transactions.Argument, 2)

	_amount, err := cadence.NewUFix64(amount)
	if err != nil {
		return &transactions.EmptyTransaction, err
	}

	aa[0] = _amount
	aa[1] = cadence.NewAddress(flow.HexToAddress(recipientAddress))

	t, err := s.Create(ctx, senderAddress, c, aa, transactions.Withdrawal)
	if err != nil {
		return t, err
	}

	return t, nil
}
