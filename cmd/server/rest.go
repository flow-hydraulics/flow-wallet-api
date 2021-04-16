package main

import (
	"context"
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/account"
	"github.com/eqlabs/flow-nft-wallet-service/pkg/config"
	"github.com/gofiber/fiber/v2"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func main() {
	conf := config.ReadFile(".flow/flow.json")
	serviceAcc := conf.Accounts["emulator-account"]

	flowClient, err := client.New("localhost:3569", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		Prefork:       false, // TODO: Enable this in production
		CaseSensitive: true,
	})

	// Just to test
	app.Get("/latestblock", func(c *fiber.Ctx) error {
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), true)
		if err != nil {
			log.Fatal(err)
		}
		return c.SendString(latestBlock.ID.String())
	})

	// Just to test
	app.Get("/generate", func(c *fiber.Ctx) error {
		_, _, myAddress := account.CreateRandom(flowClient, serviceAcc)
		return c.SendString(myAddress.Hex())
	})

	log.Fatal(app.Listen(":3000"))
}
