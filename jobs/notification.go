package jobs

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

const SendJobStatusJobType = "send_job_status"

type NotificationConfig struct {
	jobStatusWebhookUrl *url.URL
}

func (cfg *NotificationConfig) ShouldSendJobStatus() bool {
	return cfg.jobStatusWebhookUrl != nil
}

func (cfg *NotificationConfig) SendJobStatus(content string) error {
	// Handle each job status notification endpoint separately

	if err := cfg.SendJobStatusWebhook(content); err != nil {
		return err
	}

	// if err := cfg.SendSomeOtherWay(content); err != nil {
	// 	return err
	// }

	return nil
}

func (cfg *NotificationConfig) SendJobStatusWebhook(content string) error {
	if cfg.jobStatusWebhookUrl == nil {
		// Do nothing as config has no 'jobStatusWebhookUrl'
		return nil
	}

	resp, err := http.Post(cfg.jobStatusWebhookUrl.String(), "application/json", bytes.NewBuffer([]byte(content)))
	if err != nil {
		return fmt.Errorf("error while sending to webhook: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook endpoint responded with an unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
