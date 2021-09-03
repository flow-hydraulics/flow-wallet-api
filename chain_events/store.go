package chain_events

// Store manages data regarding tokens.
type Store interface {
	LockedStatus(func(*ListenerStatus) error) error
}

type LockError struct {
	Err error
}

func (e *LockError) Error() string {
	return e.Err.Error()
}
