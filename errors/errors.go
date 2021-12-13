// Package errors provides an API for errors across the application.
package errors

import "strings"

type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

func IsChainConnectionError(err error) bool {
	// TODO: check this properly
	return strings.Contains(err.Error(), "connection refused")
}
