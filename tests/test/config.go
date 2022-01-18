package test

import (
	"os"
	"path"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/onflow/flow-go-sdk"
)

// LoadConfig loads default test config
func LoadConfig(t *testing.T) *configs.Config {
	t.Helper()
	os.Clearenv()
	defer os.Clearenv()
	os.Setenv("FLOW_WALLET_ADMIN_ADDRESS", "0xf8d6e0586b0a20c7")
	os.Setenv("FLOW_WALLET_ADMIN_PRIVATE_KEY", "91a22fbd87392b019fbe332c32695c14cf2ba5b6521476a8540228bdf1987068")
	os.Setenv("FLOW_WALLET_ACCESS_API_HOST", "localhost:3569")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY", "faae4ed1c30f4e4555ee3a71f1044a8e")
	os.Setenv("FLOW_WALLET_ENCRYPTION_KEY_TYPE", "local")

	cfg, err := configs.Parse()
	if err != nil {
		t.Fatal(err)
	}

	// Check if using default
	if cfg.DatabaseDSN == "wallet.db" {
		cfg.DatabaseDSN = path.Join(t.TempDir(), "test.db")
		cfg.DatabaseType = "sqlite"
	}

	cfg.ChainID = flow.Emulator

	cfg.EnabledTokens = []string{"FUSD:0xf8d6e0586b0a20c7:fusd", "FlowToken:0x0ae53cb6e3f42a79:flowToken"}
	cfg.AdminProposalKeyCount = 100

	return cfg
}
