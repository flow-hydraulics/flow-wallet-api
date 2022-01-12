package flow_helpers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/onflow/flow-go-sdk"
	"google.golang.org/grpc"
)

type mockFlowClient struct {
	callCount uint
}

func (c *mockFlowClient) GetTransactionResult(ctx context.Context, txID flow.Identifier, opts ...grpc.CallOption) (*flow.TransactionResult, error) {
	c.callCount++

	status := flow.TransactionStatusPending

	if c.callCount >= 3 {
		status = flow.TransactionStatusSealed
	}

	return &flow.TransactionResult{Status: status}, nil
}

func (c *mockFlowClient) SendTransaction(ctx context.Context, tx flow.Transaction, opts ...grpc.CallOption) error {
	return nil
}

func TestAddressValidationAndFormatting(t *testing.T) {

	t.Run("ValidateAndFormatAddress formatting", func(t *testing.T) {
		o := "0x1cf0e2f2f715450"

		a, err := ValidateAddress(o, flow.Emulator)
		if err != nil {
			t.Fatal(err)
		}

		// Format manually
		m := strings.TrimPrefix(o, "0x")
		m = fmt.Sprintf("0x0%s", m)

		if o == a || m != a {
			t.Error("expected formatted address to have a 0 prefix")
		}
	})

	t.Run("ValidateAndFormatAddress validation", func(t *testing.T) {
		if _, err := ValidateAddress("not-a-valid-address", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}

		if _, err := ValidateAddress("", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}
	})
}

func TestWaitForSeal(t *testing.T) {
	t.Run("backoff", func(t *testing.T) {
		flowClient := new(mockFlowClient)
		ctx := context.Background()
		start := time.Now()

		if _, err := WaitForSeal(ctx, flowClient, flow.EmptyID, 0); err != nil {
			t.Fatalf("did not expect an error, got: %s", err)
		}

		if time.Since(start) < 500*time.Millisecond {
			t.Fatal("expected wait to take longer")
		}
	})
}
