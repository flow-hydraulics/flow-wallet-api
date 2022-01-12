package system

import (
	"time"
)

type ServiceOption func(*ServiceImpl)

func WithPauseDuration(duration time.Duration) ServiceOption {
	return func(svc *ServiceImpl) {
		svc.pauseDuration = duration
	}
}
