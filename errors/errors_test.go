package errors

import (
	"context"
	"fmt"
	"testing"

	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testNetError struct{}

func (e *testNetError) Error() string   { return "NetError" }
func (e *testNetError) Timeout() bool   { return false }
func (e *testNetError) Temporary() bool { return false }

func TestIsChainConnectionError(t *testing.T) {
	t.Run("error cases", func(t *testing.T) {
		var netErr error = &testNetError{}

		valid_errors := []error{
			netErr,
			client.RPCError{
				GRPCErr: status.Error(codes.DeadlineExceeded, "DeadlineExceeded"),
			},
			client.RPCError{
				GRPCErr: status.Error(codes.ResourceExhausted, "ResourceExhausted"),
			},
			client.RPCError{
				GRPCErr: status.Error(codes.Internal, "Internal"),
			},
			client.RPCError{
				GRPCErr: status.Error(codes.Unavailable, "Unavailable"),
			},
		}

		invalid_errors := []error{
			fmt.Errorf("not a connection error"),
		}

		for _, err := range valid_errors {
			if !IsChainConnectionError(err) {
				t.Fatalf("expected error to be a connection error, got \"%s\"", err)
			}
		}

		for _, err := range invalid_errors {
			if IsChainConnectionError(err) {
				t.Fatalf("expected error not to be a connection error, got \"%s\"", err)
			}
		}
	})

	t.Run("non existent gateway", func(t *testing.T) {
		fc, err := client.New("non-existent-address", grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		if _, err := fc.GetLatestBlock(context.Background(), true); err == nil {
			t.Fatal("expected an error")
		} else if !IsChainConnectionError(err) {
			t.Fatal("expected error to be a connection error")
		}
	})

}
