package jobs

import (
	"net/url"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/system"
	log "github.com/sirupsen/logrus"
)

type WorkerPoolOption func(*WorkerPool)

func WithJobStatusWebhook(u string, timeout time.Duration) WorkerPoolOption {
	return func(wp *WorkerPool) {
		if u == "" {
			return
		}

		valid, err := url.ParseRequestURI(u)
		if err != nil {
			panic("invalid job status webhook url")
		}

		if wp.notificationConfig == nil {
			wp.notificationConfig = &NotificationConfig{}
		}

		wp.notificationConfig.jobStatusWebhookUrl = valid
		wp.notificationConfig.jobStatusWebhookTimeout = timeout
	}
}

func WithSystemService(svc *system.Service) WorkerPoolOption {
	return func(wp *WorkerPool) {
		wp.systemService = svc
	}
}

func WithLogger(logger *log.Logger) WorkerPoolOption {
	return func(wp *WorkerPool) {
		wp.logger = logger
	}
}
