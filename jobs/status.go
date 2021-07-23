package jobs

import "strings"

//go:generate stringer -type=Status
type Status int

const (
	Unknown Status = iota
	Init
	Accepted
	NoAvailableWorkers
	QueueFull
	Error
	Complete
)

func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Status) UnmarshalText(text []byte) error {
	*s = StatusFromText(string(text))
	return nil
}

func StatusFromText(text string) Status {
	switch strings.ToLower(text) {
	default:
		return Unknown
	case "init":
		return Init
	case "accepted":
		return Accepted
	case "noavailableworkers":
		return NoAvailableWorkers
	case "queuefull":
		return QueueFull
	case "error":
		return Error
	case "complete":
		return Complete
	}
}
