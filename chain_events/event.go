package chain_events

import (
	"context"

	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type chainEventHandler interface {
	Handle(context.Context, flow.Event)
}

type chainEvent struct {
	handlers []chainEventHandler
}

var ChainEvent chainEvent // singleton of type event

// Register adds an event handler for this event
func (e *chainEvent) Register(handler chainEventHandler) {
	log.Debug("Registering Flow event handler")
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e *chainEvent) Trigger(ctx context.Context, payload flow.Event) {
	log.
		WithFields(log.Fields{"payload": payload}).
		Trace("Handling Flow event")

	if len(e.handlers) == 0 {
		log.Warn("No listeners for chain events")
	}

	for _, handler := range e.handlers {
		go handler.Handle(ctx, payload)
	}
}
