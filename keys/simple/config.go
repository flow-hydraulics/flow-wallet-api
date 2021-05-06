package simple

import (
	"github.com/caarlos0/env/v6"
	"github.com/eqlabs/flow-wallet-service/keys/google"
	"github.com/onflow/flow-go-sdk"
)

// Config struct for key manager.
type Config struct {
	AdminAccountAddress  string `env:"ADMIN_ACC_ADDRESS,required"`
	AdminAccountKeyIndex int    `env:"ADMIN_ACC_KEY_INDEX" envDefault:"0"`
	AdminAccountKeyType  string `env:"ADMIN_ACC_KEY_TYPE" envDefault:"local"`
	AdminAccountKeyValue string `env:"ADMIN_ACC_KEY_VALUE,required"`
	DefaultKeyStorage    string `env:"DEFAULT_KEY_STORAGE" envDefault:"local"`
	DefaultKeyIndex      int    `env:"DEFAULT_KEY_INDEX" envDefault:"0"`
	DefaultKeyWeight     int    `env:"DEFAULT_KEY_WEIGHT" envDefault:"-1"`
	EncryptionKey        string `env:"ENCRYPTION_KEY,required"`
}

// ParseConfig parses environment variables to a valid Config.
func ParseConfig() (cfg Config, googleCfg google.Config) {

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	if cfg.DefaultKeyWeight < 0 {
		cfg.DefaultKeyWeight = flow.AccountKeyWeightThreshold
	}

	if err := env.Parse(&googleCfg); err != nil {
		panic(err)
	}

	return
}
