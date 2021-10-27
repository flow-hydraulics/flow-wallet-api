package transactions

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/onflow/cadence"
)

func Test_AsCadence(t *testing.T) {
	testCases := []struct {
		name      string
		inputJson string
		expected  []cadence.Value
	}{
		{
			name:      "decode single string argument",
			inputJson: `[{"type":"String","value":"Hello"}]`,
			expected:  []cadence.Value{cadence.NewString("Hello")},
		},
		{
			name:      "decode single Uint64 argument",
			inputJson: `[{"type":"UInt64","value":"1"}]`,
			expected:  []cadence.Value{cadence.NewUInt64(1)},
		},
		{
			name:      "decode two arguments",
			inputJson: `[{"type":"String","value":"Hello"},{"type":"UInt64","value":"1"}]`,
			expected:  []cadence.Value{cadence.NewString("Hello"), cadence.NewUInt64(1)},
		},
		{
			name:      "decode empty array",
			inputJson: `[]`,
			expected:  nil,
		},
		{
			name:      "decode nil array",
			inputJson: `null`,
			expected:  nil,
		},
		{
			name:      "decode void argument",
			inputJson: `[{"type":"Void"}]`,
			expected:  []cadence.Value{cadence.NewVoid()},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			var arguments []Argument

			err := json.Unmarshal([]byte(tc.inputJson), &arguments)
			if err != nil {
				t.Fatal(err)
			}

			var decoded []cadence.Value
			for j, arg := range arguments {
				v, err := ArgAsCadence(&arg)
				if err != nil {
					t.Fatalf("j: %d, err: %#v", j, err)
				}

				decoded = append(decoded, v)
			}

			if !cmp.Equal(decoded, tc.expected) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expected, decoded))
			}
		})
	}
}
