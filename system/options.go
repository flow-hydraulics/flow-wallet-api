package system

import (
	"time"
)

type ServiceOption func(*Service)

func WithPauseDuration(duration time.Duration) ServiceOption {
	return func(svc *Service) {
		svc.pauseDuration = duration
	}
}
