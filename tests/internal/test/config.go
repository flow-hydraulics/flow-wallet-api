package test

import (
	"path"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/onflow/flow-go-sdk"
)

var defaultConfig = "default.test.env.cfg"

// LoadConfig accepts an optional config file to be loaded. If none provided,
// it loads a default one.
//
// DatabaseType is always `sqlite`.
//
// Configured database DSN points to a file in tempdir created for given test
// and it's automatically cleaned up by t.CleanUp() in the end of test run.
//
func LoadConfig(t *testing.T, cfgFile ...string) *configs.Config {
	t.Helper()

	opts := &configs.Options{EnvFilePath: defaultConfig, Version: "test"}

	// Allow optional override of config file
	if len(cfgFile) == 1 {
		opts.EnvFilePath = cfgFile[0]
	} else if len(cfgFile) > 1 {
		t.Fatalf("maximum of one config file allowed, got %d", len(cfgFile))
	}

	cfg, err := configs.ParseConfig(opts)
	if err != nil {
		t.Fatal(err)
	}

	cfg.DatabaseDSN = path.Join(t.TempDir(), "test.db")
	cfg.DatabaseType = "sqlite"
	cfg.ChainID = flow.Emulator

	return cfg
}
