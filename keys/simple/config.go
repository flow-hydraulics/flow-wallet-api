package simple

import (
	"github.com/caarlos0/env/v6"
	"github.com/onflow/flow-go-sdk"
)

// Config struct for key manager.
type Config struct {
	AdminAccountAddress  string `env:"ADMIN_ADDRESS,required"`
	AdminAccountKeyIndex int    `env:"ADMIN_KEY_INDEX" envDefault:"0"`
	AdminAccountKeyType  string `env:"ADMIN_KEY_TYPE" envDefault:"local"`
	AdminAccountKeyValue string `env:"ADMIN_PRIVATE_KEY,required"`
	DefaultKeyType       string `env:"DEFAULT_KEY_TYPE" envDefault:"local"`
	DefaultKeyIndex      int    `env:"DEFAULT_KEY_INDEX" envDefault:"0"`
	DefaultKeyWeight     int    `env:"DEFAULT_KEY_WEIGHT" envDefault:"-1"`
	DefaultSignAlgo      string `env:"DEFAULT_SIGN_ALGO" envDefault:"ECDSA_P256"`
	DefaultHashAlgo      string `env:"DEFAULT_HASH_ALGO" envDefault:"SHA3_256"`
	EncryptionKey        string `env:"ENCRYPTION_KEY,required"`
}

// ParseConfig parses environment variables to a valid Config.
func ParseConfig() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	if cfg.DefaultKeyWeight < 0 {
		cfg.DefaultKeyWeight = flow.AccountKeyWeightThreshold
	}

	return
}
