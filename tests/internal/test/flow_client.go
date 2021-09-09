package test

import (
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func NewFlowClient(t *testing.T, cfg *configs.Config) *client.Client {
	fc, err := client.New(cfg.AccessAPIHost, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	close := func() {
		err := fc.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Cleanup(close)

	return fc
}
