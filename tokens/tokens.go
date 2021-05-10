// Package tokens provides functions for token handling in Flow blockhain.
package tokens

import "path/filepath"

type Token struct {
	Name        string
	BaseAddress string
	Address     string
}

func TemplatePath() string {
	// TODO: figure out a dynamic way to do this?
	return filepath.Join("cadence")
}
