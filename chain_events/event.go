package chain_events

import (
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type handler interface {
	Handle(flow.Event)
}

type event struct {
	handlers []handler
}

var Event event // singleton of type event

// Register adds an event handler for this event
func (e *event) Register(handler handler) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e *event) Trigger(payload flow.Event) {
	if len(e.handlers) == 0 {
		log.Warn("No listeners for chain events")
	}

	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
