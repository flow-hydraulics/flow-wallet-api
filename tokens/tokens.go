// Package tokens provides functions for token handling in Flow blockhain.
// https://docs.onflow.org/core-contracts
package tokens

type TokenDetails struct {
	Name    string `json:"name"`
	Balance string `json:"balance"`
}
