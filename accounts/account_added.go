package accounts

import (
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type AccountAddedPayload struct {
	Address                   flow.Address
	InitializedFungibleTokens []templates.Token
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
	log.Debug("Registering AccountAdded event handler")
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e *accountAdded) Trigger(payload AccountAddedPayload) {
	log.
		WithFields(log.Fields{"payload": payload}).
		Trace("Handling AccountAdded event")

	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
