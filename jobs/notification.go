package jobs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const SendJobStatusJobType = "send_job_status"

type NotificationConfig struct {
	jobStatusWebhookUrl     *url.URL
	jobStatusWebhookTimeout time.Duration
}

func (cfg *NotificationConfig) ShouldSendJobStatus() bool {
	return cfg.jobStatusWebhookUrl != nil
}

func (cfg *NotificationConfig) SendJobStatus(ctx context.Context, content string) error {
	// Handle each job status notification endpoint separately

	if err := cfg.SendJobStatusWebhook(ctx, content); err != nil {
		return err
	}

	// if err := cfg.SendSomeOtherWay(content); err != nil {
	// 	return err
	// }

	return nil
}

func (cfg *NotificationConfig) SendJobStatusWebhook(ctx context.Context, content string) error {
	if cfg.jobStatusWebhookUrl == nil {
		// Do nothing as config has no 'jobStatusWebhookUrl'
		return nil
	}

	client := http.Client{
		Timeout: cfg.jobStatusWebhookTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.jobStatusWebhookUrl.String(), bytes.NewBuffer([]byte(content)))
	if err != nil {
		return fmt.Errorf("error while creating webhook request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending webhook request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook endpoint responded with an unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
