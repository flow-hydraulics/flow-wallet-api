// Package errors provides an API for errors across the application.
package errors

import (
	"net"

	"github.com/onflow/flow-go-sdk/access/grpc"
	"google.golang.org/grpc/codes"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

var accessAPIConnectionErrors = []codes.Code{
	codes.DeadlineExceeded,
	codes.ResourceExhausted,
	codes.Internal,
	codes.Unavailable,
}

func IsChainConnectionError(err error) bool {
	if _, ok := err.(net.Error); ok {
		return true
	}

	if err, ok := err.(grpc.RPCError); ok {
		// Check for Flow Access API connection errors
		for _, code := range accessAPIConnectionErrors {
			if err.GRPCStatus().Code() == code {
				return true
			}
		}
	}
	return false
}
