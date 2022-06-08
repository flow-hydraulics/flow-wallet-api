package internal

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	access "github.com/onflow/flow-go-sdk/access/grpc"
	"google.golang.org/grpc"
)

type MockFlowClient struct {
	getTransactionResultCallCount uint
}

func (c *MockFlowClient) ExecuteScriptAtLatestBlock(ctx context.Context, script []byte, arguments []cadence.Value, opts ...grpc.CallOption) (cadence.Value, error) {
	return nil, nil
}

func (c *MockFlowClient) GetAccount(ctx context.Context, address flow.Address, opts ...grpc.CallOption) (*flow.Account, error) {
	return nil, nil
}

func (c *MockFlowClient) GetAccountAtLatestBlock(ctx context.Context, address flow.Address, opts ...grpc.CallOption) (*flow.Account, error) {
	return nil, nil
}

func (c *MockFlowClient) GetTransaction(ctx context.Context, txID flow.Identifier, opts ...grpc.CallOption) (*flow.Transaction, error) {
	return nil, nil
}

func (c *MockFlowClient) GetTransactionResult(ctx context.Context, txID flow.Identifier, opts ...grpc.CallOption) (*flow.TransactionResult, error) {
	c.getTransactionResultCallCount++
	status := flow.TransactionStatusPending
	if c.getTransactionResultCallCount >= 3 {
		status = flow.TransactionStatusSealed
	}
	return &flow.TransactionResult{Status: status}, nil
}

func (c *MockFlowClient) GetLatestBlockHeader(ctx context.Context, isSealed bool, opts ...grpc.CallOption) (*flow.BlockHeader, error) {
	return nil, nil
}

func (c *MockFlowClient) GetEventsForHeightRange(ctx context.Context, query access.EventRangeQuery, opts ...grpc.CallOption) ([]client.BlockEvents, error) {
	return nil, nil
}

func (c *MockFlowClient) SendTransaction(ctx context.Context, tx flow.Transaction, opts ...grpc.CallOption) error {
	return nil
}
