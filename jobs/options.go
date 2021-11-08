package jobs

import "net/url"

type WorkerPoolOption func(*WorkerPool)

func WithJobStatusWebHook(u string) WorkerPoolOption {
	return func(wp *WorkerPool) {
		if u == "" {
			return
		}

		valid, err := url.ParseRequestURI(u)
		if err != nil {
			panic(err)
		}

		if wp.notificationConfig == nil {
			wp.notificationConfig = &NotificationConfig{}
		}

		wp.notificationConfig.jobStatusWebHookUrl = valid
	}
}
