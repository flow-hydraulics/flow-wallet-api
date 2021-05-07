package account

import (
	"testing"
)

func TestAccountService(t *testing.T) {
	service := &Service{cfg: Config{ChainId: "flow-emulator"}}

	t.Run("ValidateAddress", func(t *testing.T) {
		if err := service.ValidateAddress("not-a-valid-address"); err == nil {
			t.Error("Expected an error")
		}

		if err := service.ValidateAddress(""); err == nil {
			t.Error("Expected an error")
		}
	})
}
