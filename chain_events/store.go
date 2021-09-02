package chain_events

// Store manages data regarding tokens.
type Store interface {
	LockedStatus(func(*ListenerStatus) error) error
}
