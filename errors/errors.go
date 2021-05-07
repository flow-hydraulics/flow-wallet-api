// Package errors provides an API for errors across the application.
package errors

type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

type JobQueueFull struct {
	Err error
}

func (e *JobQueueFull) Error() string {
	return e.Err.Error()
}
