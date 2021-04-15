package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

func main() {

	flowClient, err := client.New("localhost:3569", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		Prefork:       false, // TODO: Enable this in production
		CaseSensitive: true,
	})

	// Just to test the networking with the emulator
	app.Get("/test", func(c *fiber.Ctx) error {
		latestBlock, err := flowClient.GetLatestBlock(context.Background(), true)
		if err != nil {
			log.Fatal(err)
		}
		return c.SendString(latestBlock.ID.String())
	})

	log.Fatal(app.Listen(":3000"))
}
