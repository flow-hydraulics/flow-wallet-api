package chain_events

// Store manages data regarding tokens.
type Store interface {
	UpdateListenerStatus(s *ListenerStatus) error
	GetListenerStatus() (*ListenerStatus, error)
}
