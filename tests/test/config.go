package test

import (
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
)

// LoadConfig loads test config
func LoadConfig(t *testing.T) *configs.Config {
	cfg := configs.ParseTestConfig(t)
	configs.ConfigureLogger(cfg.LogLevel)
	return cfg
}
