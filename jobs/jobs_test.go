package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/google/uuid"
)

type dummyStore struct{}

func (*dummyStore) Jobs(datastore.ListOptions) ([]Job, error) { return nil, nil }
func (*dummyStore) Job(id uuid.UUID) (Job, error)             { return Job{}, nil }
func (*dummyStore) InsertJob(*Job) error                      { return nil }
func (*dummyStore) UpdateJob(*Job) error                      { return nil }
func (*dummyStore) IncreaseExecCount(j *Job) error            { return nil }
func (*dummyStore) SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod time.Duration, o datastore.ListOptions) ([]Job, error) {
	return nil, nil
}
func (*dummyStore) Status() ([]StatusQuery, error) { return nil, nil }

type dummyWriter struct {
	T      *testing.T
	record []string
}

func (writer *dummyWriter) Write(p []byte) (n int, err error) {
	writer.T.Helper()
	writer.record = append(writer.record, string(p))
	return 0, nil
}

func TestScheduleSendNotification(t *testing.T) {
	writer := &dummyWriter{T: t}
	logger := log.New()
	logger.Out = writer

	ctx, cancel := context.WithCancel(context.Background())
	wp := WorkerPool{
		context:       ctx,
		cancelContext: cancel,
		executors:     make(map[string]ExecutorFunc),
		jobChan:       make(chan *Job, 1),
		logger:        log.NewEntry(logger),
		store:         &dummyStore{},
	}

	WithJobStatusWebhook("http://localhost", time.Minute)(&wp)

	sendNotificationCalled := false

	wp.RegisterExecutor(SendJobStatusJobType, func(ctx context.Context, j *Job) error {
		j.ShouldSendNotification = false
		sendNotificationCalled = true
		return nil
	})

	wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
		j.ShouldSendNotification = true
		return nil
	})

	job, err := wp.CreateJob("TestJobType", "")
	if err != nil {
		t.Fatal(err)
	}

	wp.process(job)

	if len(wp.jobChan) == 0 {
		t.Fatal("expected job channel to contain a job")
	}

	sendNotificationJob := <-wp.jobChan

	if sendNotificationJob.Type != "send_job_status" {
		t.Fatalf("expected pool to have a send_job_status job")
	}

	wp.process(sendNotificationJob)

	if !sendNotificationCalled {
		t.Fatalf("expected 'sendNotificationCalled' to equal true")
	}

	if len(writer.record) > 0 {
		t.Fatalf("did not expect a warning, got %s", writer.record)
	}
}

func TestExecuteSendNotification(t *testing.T) {
	t.Run("valid job should send", func(t *testing.T) {
		var webhookJob Job
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&webhookJob); err != nil {
				t.Fatal(err)
			}
		}))
		defer svr.Close()

		writer := &dummyWriter{T: t}
		logger := log.New()
		logger.Out = writer

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPool{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			logger:        log.NewEntry(logger),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return nil
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		wp.process(job)
		wp.process(<-wp.jobChan)

		if webhookJob.Type != "TestJobType" {
			t.Fatalf("expected webhook endpoint to have received a notification")
		}

		if webhookJob.State != Complete {
			t.Fatalf("expected job to be in state '%s' got '%s'", Complete, webhookJob.State)
		}

		if len(writer.record) > 0 {
			t.Fatalf("did not expect a warning, got %s", writer.record)
		}
	})

	t.Run("failed job should send", func(t *testing.T) {
		var webhookJob Job
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(&webhookJob); err != nil {
				t.Fatal(err)
			}
		}))
		defer svr.Close()

		writer := &dummyWriter{T: t}
		logger := log.New()
		logger.Out = writer

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPool{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			logger:        log.NewEntry(logger),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return ErrPermanentFailure
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		wp.process(job)
		wp.process(<-wp.jobChan)

		if webhookJob.Type != "TestJobType" {
			t.Fatalf("expected webhook endpoint to have received a notification")
		}

		if webhookJob.State != Failed {
			t.Fatalf("expected job to be in state '%s' got '%s'", Failed, webhookJob.State)
		}

		if len(writer.record) == 0 {
			t.Fatalf("expected a warning, got %s", writer.record)
		}
	})

	t.Run("erroring job should not send", func(t *testing.T) {
		writer := &dummyWriter{T: t}
		logger := log.New()
		logger.Out = writer

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPool{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			logger:        log.NewEntry(logger),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook("http://localhost", time.Minute)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return fmt.Errorf("test error")
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		wp.process(job)

		if len(wp.jobChan) != 0 {
			t.Errorf("did not expect a job to be queued")
		}
	})

	t.Run("valid job should send but get an endpoint error and retry send", func(t *testing.T) {
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "test error", http.StatusBadGateway)
		}))
		defer svr.Close()

		writer := &dummyWriter{T: t}
		logger := log.New()
		logger.Out = writer

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPool{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			logger:        log.NewEntry(logger),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return nil
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		wp.process(job)
		sendNotificationJob := <-wp.jobChan
		wp.process(sendNotificationJob)

		if len(writer.record) != 1 {
			t.Errorf("expected there to be one warning, got %s", writer.record)
		}

		if sendNotificationJob.State != Error {
			t.Errorf("expected send notification job to be in '%s' state, got '%s'", Error, sendNotificationJob.State)
		}
	})
}
