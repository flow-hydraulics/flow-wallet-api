package account

import (
	"context"
	"testing"

	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

var serviceAcct Account = ReadFromFlowFile("../../flow.json", "emulator-account")

var testNft NFT = NFT{
	BaseAddress: serviceAcct.Address,
	Address:     serviceAcct.Address,
	Name:        "KittyItems",
}

func TestSetupNFT(t *testing.T) {
	ctx := context.Background()

	c, err := client.New("localhost:3569", grpc.WithInsecure())
	if err != nil {
		t.Error("Error while initialising flow client")
		return
	}

	newAcc := CreateRandom(ctx, c, serviceAcct)
	newAcc.SetupNFT(ctx, c, serviceAcct, testNft)
}
