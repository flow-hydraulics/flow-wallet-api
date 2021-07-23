package jobs

import (
	"encoding/json"
	"testing"
)

func TestStatus(t *testing.T) {

	cases := []struct {
		name   string
		status Status
		json   string
	}{
		{
			name:   `"Unknown"`,
			status: Unknown,
			json:   `"somethingunknown"`,
		},
		{
			name:   `"Init"`,
			status: Init,
			json:   `"init"`,
		},
		{
			name:   `"Accepted"`,
			status: Accepted,
			json:   `"accepted"`,
		},
		{
			name:   `"NoAvailableWorkers"`,
			status: NoAvailableWorkers,
			json:   `"noavailableworkers"`,
		},
		{
			name:   `"QueueFull"`,
			status: QueueFull,
			json:   `"queuefull"`,
		},
		{
			name:   `"Error"`,
			status: Error,
			json:   `"error"`,
		},
		{
			name:   `"Complete"`,
			status: Complete,
			json:   `"complete"`,
		},
	}
	for _, c := range cases {
		// Marshalling
		t.Run(c.name+" marshal", func(t *testing.T) {
			b, err := json.Marshal(c.status)
			if err != nil {
				t.Error(err)
			}
			if string(b) != c.name {
				t.Errorf("expected %v got %v", c.name, string(b))
			}
		})

		// UnMarshalling
		t.Run(c.name+" unmarshal", func(t *testing.T) {
			var s Status
			if err := json.Unmarshal([]byte(c.json), &s); err != nil {
				t.Error(err)
			}
			if s != c.status {
				t.Errorf("expected %v got %v", c.status, s)
			}
		})
	}
}
