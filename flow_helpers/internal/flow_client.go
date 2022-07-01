package internal

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

type MockFlowClient struct {
	getTransactionResultCallCount uint
}

func (c *MockFlowClient) ExecuteScriptAtLatestBlock(ctx context.Context, script []byte, arguments []cadence.Value) (cadence.Value, error) {
	return nil, nil
}

func (c *MockFlowClient) GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error) {
	return nil, nil
}

func (c *MockFlowClient) GetAccountAtLatestBlock(ctx context.Context, address flow.Address) (*flow.Account, error) {
	return nil, nil
}

func (c *MockFlowClient) GetTransaction(ctx context.Context, txID flow.Identifier) (*flow.Transaction, error) {
	return nil, nil
}

func (c *MockFlowClient) GetTransactionResult(ctx context.Context, txID flow.Identifier) (*flow.TransactionResult, error) {
	c.getTransactionResultCallCount++
	status := flow.TransactionStatusPending
	if c.getTransactionResultCallCount >= 3 {
		status = flow.TransactionStatusSealed
	}
	return &flow.TransactionResult{Status: status}, nil
}

func (c *MockFlowClient) GetLatestBlockHeader(ctx context.Context, isSealed bool) (*flow.BlockHeader, error) {
	return nil, nil
}

func (c *MockFlowClient) GetEventsForHeightRange(ctx context.Context, eventType string, startHeight uint64, endHeight uint64) ([]flow.BlockEvents, error) {
	return nil, nil
}

func (c *MockFlowClient) SendTransaction(ctx context.Context, tx flow.Transaction) error {
	return nil
}
