package events

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

var ChainEvent chainEvent

type chainEvent struct {
	handlers []interface{ Handle(flow.Event) }
}

// Register adds an event handler for this event
func (e *chainEvent) Register(handler interface{ Handle(flow.Event) }) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e chainEvent) Trigger(payload flow.Event) {
	if len(e.handlers) == 0 {
		fmt.Println("Warning: no listeners for chain events")
	}
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
