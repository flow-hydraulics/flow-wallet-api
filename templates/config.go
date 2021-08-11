package templates

import (
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/onflow/flow-go-sdk"
)

// config struct for templates.
type config struct {
	ChainID          flow.ChainID `env:"CHAIN_ID" envDefault:"flow-emulator"`
	EnvEnabledTokens []string     `env:"ENABLED_TOKENS" envSeparator:","`
	enabledTokens    map[string]Token
}

var cachedConfig *config

// parseConfig parses environment variables to a valid Config.
func parseConfig() *config {
	if cachedConfig == nil {
		cachedConfig = &config{}

		if err := env.Parse(cachedConfig); err != nil {
			panic(err)
		}

		cachedConfig.enabledTokens = make(map[string]Token, len(cachedConfig.EnvEnabledTokens))
		for _, s := range cachedConfig.EnvEnabledTokens {
			ss := strings.Split(s, ":")
			token := Token{Name: ss[0], Address: ss[1]}
			if len(ss) > 2 {
				token.NameLowerCase = ss[2]
			}
			// Use all lowercase as the key so we can do case insenstive matchig
			// in URLs
			key := strings.ToLower(ss[0])
			cachedConfig.enabledTokens[key] = token
		}
	}

	return cachedConfig
}
