package flow_helpers

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
)

func TestValidateAddress(t *testing.T) {
	t.Run("ValidateAddress", func(t *testing.T) {
		if err := ValidateAddress("not-a-valid-address", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}

		if err := ValidateAddress("", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}
	})
}
