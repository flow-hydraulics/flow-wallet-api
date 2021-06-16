package templates

import (
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/onflow/flow-go-sdk"
)

// config struct for templates.
type config struct {
	ChainId               flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
	EnvEnabledTokens      []string     `env:"ENABLED_TOKENS" envSeparator:","`
	enabledTokens         []Token
	enabledTokenAddresses map[string]string
}

var cachedConfig *config

// parseConfig parses environment variables to a valid Config.
func parseConfig() *config {
	if cachedConfig == nil {
		cachedConfig = &config{}

		if err := env.Parse(cachedConfig); err != nil {
			panic(err)
		}

		cachedConfig.enabledTokens = make([]Token, len(cachedConfig.EnvEnabledTokens))
		cachedConfig.enabledTokenAddresses = make(map[string]string, len(cachedConfig.EnvEnabledTokens))
		for i, s := range cachedConfig.EnvEnabledTokens {
			ss := strings.Split(s, ":")
			cachedConfig.enabledTokens[i] = Token{Name: ss[0], Address: ss[1]}
			cachedConfig.enabledTokenAddresses[ss[0]] = ss[1]
		}
	}

	return cachedConfig
}
