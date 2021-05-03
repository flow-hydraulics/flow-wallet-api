package account

import (
	"github.com/caarlos0/env/v6"
	"github.com/onflow/flow-go-sdk"
)

type Config struct {
	ChainId flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
}

func ParseConfig() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return
}
