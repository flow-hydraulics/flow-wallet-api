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
	"github.com/jpillora/backoff"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

type FlowClient interface {
	ExecuteScriptAtLatestBlock(ctx context.Context, script []byte, arguments []cadence.Value) (cadence.Value, error)
	GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error)
	GetAccountAtLatestBlock(ctx context.Context, address flow.Address) (*flow.Account, error)
	GetTransaction(ctx context.Context, txID flow.Identifier) (*flow.Transaction, error)
	GetTransactionResult(ctx context.Context, txID flow.Identifier) (*flow.TransactionResult, error)
	GetLatestBlockHeader(ctx context.Context, isSealed bool) (*flow.BlockHeader, error)
	GetEventsForHeightRange(ctx context.Context, eventType string, startHeight uint64, endHeight uint64) ([]flow.BlockEvents, error)
	SendTransaction(ctx context.Context, tx flow.Transaction) error
}

const hexPrefix = "0x"

// LatestBlockId retuns the flow.Identifier for the latest block in the chain.
func LatestBlockId(ctx context.Context, flowClient FlowClient) (*flow.Identifier, error) {
	block, err := flowClient.GetLatestBlockHeader(ctx, false)
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
func WaitForSeal(ctx context.Context, flowClient FlowClient, id flow.Identifier, timeout time.Duration) (*flow.TransactionResult, error) {
	var (
		result *flow.TransactionResult
		err    error
	)

	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    time.Second,
		Factor: 5,
		Jitter: true,
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		result, err = flowClient.GetTransactionResult(ctx, id)
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

		time.Sleep(b.Duration())
	}
}

// SendAndWait sends the transaction and waits for the transaction to be sealed
func SendAndWait(ctx context.Context, flowClient FlowClient, tx flow.Transaction, timeout time.Duration) (*flow.TransactionResult, error) {
	if err := flowClient.SendTransaction(ctx, tx); err != nil {
		return nil, err
	}
	return WaitForSeal(ctx, flowClient, tx.ID(), timeout)
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
