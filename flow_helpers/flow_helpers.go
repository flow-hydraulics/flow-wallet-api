// Package flow_helpers provides convience functions for Flow blockchain interaction.
package flow_helpers

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// LatestBlockId retuns the flow.Identifier for the latest block in the chain.
func LatestBlockId(ctx context.Context, c *client.Client) (flow.Identifier, error) {
	block, err := c.GetLatestBlock(ctx, true)
	if err != nil {
		return flow.Identifier{}, err
	}
	return block.ID, nil
}

// WaitForSeal blocks until either an error occurs or the transaction
// identified by "id" gets a "TransactionStatusSealed" status.
func WaitForSeal(ctx context.Context, c *client.Client, id flow.Identifier) (result *flow.TransactionResult, err error) {
	result, err = c.GetTransactionResult(ctx, id)
	if err != nil {
		return
	}

	if result.Error != nil {
		err = result.Error
		return
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, id)

		if err != nil {
			return
		}

		if result.Error != nil {
			err = result.Error
			return
		}
	}

	return result, nil
}

// ValidateAddress checks if the given address is valid in the current Flow network.
func ValidateAddress(address string, chainId flow.ChainID) (err error) {
	flowAddress := flow.HexToAddress(address)
	if !flowAddress.IsValid(chainId) {
		err = &errors.RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("not a valid address"),
		}
	}

	return
}

func ValidateTransactionId(id string) error {
	invalidErr := &errors.RequestError{
		StatusCode: http.StatusBadRequest,
		Err:        fmt.Errorf("not a valid transaction id"),
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
