package account

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/onflow/flow-go-sdk"
)

type Config struct {
	ChainId flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
}

var cfg Config

func init() {
	godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}
