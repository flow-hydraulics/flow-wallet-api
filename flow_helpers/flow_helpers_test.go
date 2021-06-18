package flow_helpers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
)

func TestAddressValidationAndFormatting(t *testing.T) {

	t.Run("ValidateAndFormatAddress formatting", func(t *testing.T) {
		o := "0x1cf0e2f2f715450"

		a, err := ValidateAddress(o, flow.Emulator)
		if err != nil {
			t.Fatal(err)
		}

		// Format manually
		m := strings.TrimPrefix(o, "0x")
		m = fmt.Sprintf("0x0%s", m)

		if o == a || m != a {
			t.Error("expected formatted address to have a 0 prefix")
		}
	})

	t.Run("ValidateAndFormatAddress validation", func(t *testing.T) {
		if _, err := ValidateAddress("not-a-valid-address", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}

		if _, err := ValidateAddress("", flow.Emulator); err == nil {
			t.Error("Expected an error")
		}
	})
}
