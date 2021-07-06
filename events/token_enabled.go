package events

var TokenEnabled tokenEnabled

// TokenEnabledPayload is the data for when a token is enabled
type TokenEnabledPayload struct {
	TokenName    string
	TokenAddress string
	TokenType    string
}

type tokenEnabled struct {
	handlers []interface{ Handle(TokenEnabledPayload) }
}

// Register adds an event handler for this event
func (e *tokenEnabled) Register(handler interface{ Handle(TokenEnabledPayload) }) {
	e.handlers = append(e.handlers, handler)
}

// Trigger sends out an event with the payload
func (e tokenEnabled) Trigger(payload TokenEnabledPayload) {
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}
