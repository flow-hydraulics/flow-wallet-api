package jobs

import "net/url"

type WorkerPoolOption func(*WorkerPool)

func WithJobStatusWebhook(u string) WorkerPoolOption {
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
	}
}
