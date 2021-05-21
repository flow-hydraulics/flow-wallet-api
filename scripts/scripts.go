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

func (s *Script) MustDecodeArgs() []cadence.Value {
	var aa []cadence.Value

	for _, sa := range s.Arguments {
		a, err := DecodeArgument(sa)
		if err != nil {
			panic("unable to decode arguments")
		}
		aa = append(aa, a)
	}

	return aa
}

func DecodeArgument(a Argument) (cadence.Value, error) {
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
