// Package flow_helpers provides convience functions for Flow blockchain interaction.
package flow_helpers

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

type GetTransactionResultFunc func(ctx context.Context, id flow.Identifier, opts ...grpc.CallOption) (*flow.TransactionResult, error)

const hexPrefix = "0x"

// LatestBlockId retuns the flow.Identifier for the latest block in the chain.
func LatestBlockId(ctx context.Context, c *client.Client) (*flow.Identifier, error) {
	block, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		return nil, err
	}
	return &block.ID, nil
}

// WaitForSeal blocks until
// - an error occurs while fetching the transaction result
// - the transaction gets an error status
// - the transaction gets a "TransactionStatusSealed" or "TransactionStatusExpired" status
// - timeout is reached
func WaitForSeal(ctx context.Context, getResult GetTransactionResultFunc, id flow.Identifier, timeout time.Duration) (*flow.TransactionResult, error) {
	var (
		result *flow.TransactionResult
		err    error
	)

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		result, err = getResult(ctx, id)
		if err != nil {
			return nil, err
		}

		if result.Error != nil {
			return result, result.Error
		}

		switch result.Status {
		default:
			// Not an interesting state, exit switch and continue loop
		case flow.TransactionStatusExpired:
			// Expired, handle as an error
			return result, fmt.Errorf("transaction expired")
		case flow.TransactionStatusSealed:
			// Sealed, all good
			return result, nil
		}

		time.Sleep(time.Second)
	}
}

// SendAndWait sends the transaction and waits for the transaction to be sealed
func SendAndWait(ctx context.Context, c *client.Client, tx flow.Transaction, timeout time.Duration) (*flow.TransactionResult, error) {
	if err := c.SendTransaction(ctx, tx); err != nil {
		return nil, err
	}
	return WaitForSeal(ctx, c, tx.ID(), timeout)
}

func HexString(str string) string {
	if strings.HasPrefix(str, hexPrefix) {
		return str
	}
	return fmt.Sprintf("%s%s", hexPrefix, str)
}

func FormatAddress(address flow.Address) string {
	return HexString(address.Hex())
}

func ValidateAddress(address string, chainId flow.ChainID) (string, error) {
	flowAddress := flow.HexToAddress(address)
	if !flowAddress.IsValid(chainId) {
		return "", &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf(`not a valid address: "%s"`, address),
		}
	}
	return FormatAddress(flowAddress), nil
}

func ValidateTransactionId(id string) error {
	invalidErr := &errors.RequestError{
		StatusCode: http.StatusBadRequest,
		Err:        fmt.Errorf(`not a valid transaction id: "%s"`, id),
	}
	b, err := hex.DecodeString(id)
	if err != nil {
		return invalidErr
	}
	if id != flow.BytesToID(b).Hex() {
		return invalidErr
	}
	return nil
}
