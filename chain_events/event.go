package chain_events

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
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
		fmt.Println("Warning: no listeners for chain events")
	}
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
