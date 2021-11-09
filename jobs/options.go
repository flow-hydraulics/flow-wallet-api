package jobs

import (
	"net/url"
	"time"
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
