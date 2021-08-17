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
)

const hexPrefix = "0x"

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
func WaitForSeal(ctx context.Context, c *client.Client, id flow.Identifier) (*flow.TransactionResult, error) {
	result, err := c.GetTransactionResult(ctx, id)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		result, err = c.GetTransactionResult(ctx, id)
		if err != nil {
			return nil, err
		}

		if result.Error != nil {
			return nil, result.Error
		}
	}

	return result, nil
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
