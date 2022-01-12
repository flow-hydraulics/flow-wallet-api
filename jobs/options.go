package jobs

import (
	"net/url"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/system"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

type WorkerPoolOption func(*WorkerPoolImpl)
type JobOption func(*Job)

func WithJobStatusWebhook(u string, timeout time.Duration) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
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

func WithSystemService(svc system.Service) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.systemService = svc
	}
}

func WithLogger(logger *log.Logger) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.logger = logger
	}
}

func WithMaxJobErrorCount(count int) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.maxJobErrorCount = count
	}
}

func WithDbJobPollInterval(d time.Duration) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.dbJobPollInterval = d
	}
}

func WithAcceptedGracePeriod(d time.Duration) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.acceptedGracePeriod = d
	}
}

func WithReSchedulableGracePeriod(d time.Duration) WorkerPoolOption {
	return func(wp *WorkerPoolImpl) {
		wp.reSchedulableGracePeriod = d
	}
}

func WithAttributes(attributes datatypes.JSON) JobOption {
	return func(job *Job) {
		job.Attributes = attributes
	}
}
