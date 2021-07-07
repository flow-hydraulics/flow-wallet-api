package accounts

import (
	"github.com/onflow/flow-go-sdk"
)

type AccountAddedPayload struct {
	Address flow.Address
}

type accountAddedHandler interface {
	Handle(AccountAddedPayload)
}

type accountAdded struct {
	handlers []accountAddedHandler
}

var AccountAdded accountAdded // singleton of type accountAdded

// Register adds an event handler for this event
func (e *accountAdded) Register(handler accountAddedHandler) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e *accountAdded) Trigger(payload AccountAddedPayload) {
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
