package configs

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Setenv("FLOW_WALLET_ENABLED_TOKENS", "FlowToken:0x0ae53cb6e3f42a79:flowToken,FUSD:0xf8d6e0586b0a20c7:fusd")
	t.Setenv("FLOW_WALLET_ADMIN_ADDRESS", "admin-address")
	t.Setenv("FLOW_WALLET_ADMIN_PRIVATE_KEY", "admin-private-key")
	t.Setenv("FLOW_WALLET_ENCRYPTION_KEY", "encryption-key")
	t.Setenv("FLOW_WALLET_ENCRYPTION_KEY_TYPE", "local")
	t.Setenv("FLOW_WALLET_ACCESS_API_HOST", "access-api-host")
	t.Setenv("FLOW_WALLET_WORKER_COUNT", "1")

	cfg, err := Parse()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.AdminAddress != "admin-address" {
		t.Errorf(`expected "AdminAddress" to equal "admin-address", got "%s"`, cfg.AdminAddress)
	}

	if cfg.WorkerCount != 1 {
		t.Errorf(`expected "WorkerCount" to equal 1, got %d`, cfg.WorkerCount)
	}

	if len(cfg.EnabledTokens) != 2 ||
		cfg.EnabledTokens[0] != "FlowToken:0x0ae53cb6e3f42a79:flowToken" ||
		cfg.EnabledTokens[1] != "FUSD:0xf8d6e0586b0a20c7:fusd" {
		t.Errorf(
			"expected %#v, got %#v",
			[]string{"FlowToken:0x0ae53cb6e3f42a79:flowToken", "FUSD:0xf8d6e0586b0a20c7:fusd"},
			cfg.EnabledTokens,
		)
	}
}
