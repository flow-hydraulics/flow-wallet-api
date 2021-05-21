package scripts

import (
	"encoding/json"

	"github.com/onflow/cadence"
	c_json "github.com/onflow/cadence/encoding/json"
)

type Script struct {
	Code      string     `json:"code"`
	Arguments []Argument `json:"arguments"`
}

type Argument interface{}

func ArgToCadence(a Argument) (cadence.Value, error) {
	j, err := json.Marshal(a)
	if err != nil {
		return cadence.Void{}, err
	}
	c, err := c_json.Decode(j)
	if err != nil {
		return cadence.Void{}, err
	}
	return c, nil
}
