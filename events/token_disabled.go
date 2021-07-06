package events

var TokenDisabled tokenDisabled

// TokenDisabledPayload is the data for when a token is disabled
type TokenDisabledPayload struct {
	TokenName    string
	TokenAddress string
	TokenType    string
}

type tokenDisabled struct {
	handlers []interface{ Handle(TokenDisabledPayload) }
}

// Register adds an event handler for this event
func (e *tokenDisabled) Register(handler interface{ Handle(TokenDisabledPayload) }) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e tokenDisabled) Trigger(payload TokenDisabledPayload) {
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
