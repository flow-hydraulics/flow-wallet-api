package transactions

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/onflow/cadence"
	c_json "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

func AsCadence(a *Argument) (cadence.Value, error) {
	// Convert to json bytes so we can use cadence's own encoding library
	j, err := json.Marshal(a)
	if err != nil {
		return cadence.Void{}, err
	}

	// Use cadence's own encoding library
	c, err := c_json.Decode(j)
	if err != nil {
		return cadence.Void{}, err
	}

	return c, nil
}

func MustDecodeArgs(aa []Argument) []cadence.Value {
	var cc []cadence.Value

	for _, a := range aa {
		c, err := AsCadence(&a)
		if err != nil {
			panic("unable to decode arguments")
		}
		cc = append(cc, c)
	}

	return cc
}

func ValidateTransactionId(id string) error {
	invalidErr := &errors.RequestError{
		StatusCode: http.StatusBadRequest,
		Err:        fmt.Errorf("not a valid transaction id"),
	}
	b, err := hex.DecodeString(id)
	if err != nil {
		return invalidErr
	}
	if id != flow.BytesToID(b).Hex() {
		return invalidErr
	}
	return nil
}
