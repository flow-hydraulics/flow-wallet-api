// Package errors provides an API for errors across the application.
package errors

type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}
