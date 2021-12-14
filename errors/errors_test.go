package errors

import (
	"context"
	"testing"

	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func TestIsChainConnectionError(t *testing.T) {
	fc, err := client.New("non-existent-address", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fc.GetLatestBlock(context.Background(), true); err == nil {
		t.Fatal("expected an error")
	} else if !IsChainConnectionError(err) {
		t.Fatal("expected error to be a connection error")
	}
}
