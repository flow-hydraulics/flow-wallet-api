package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/google/uuid"
)

type dummyStore struct{}

func (*dummyStore) Jobs(datastore.ListOptions) ([]Job, error) { return nil, nil }
func (*dummyStore) Job(id uuid.UUID) (Job, error)             { return Job{}, nil }
func (*dummyStore) InsertJob(*Job) error                      { return nil }
func (*dummyStore) UpdateJob(*Job) error                      { return nil }
func (*dummyStore) AcceptJob(j *Job, acceptedGracePeriod time.Duration) error {
	j.ExecCount = j.ExecCount + 1
	return nil
}
func (*dummyStore) SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod time.Duration, o datastore.ListOptions) ([]Job, error) {
	return nil, nil
}
func (*dummyStore) Status() ([]StatusQuery, error) { return nil, nil }

func TestScheduleSendNotification(t *testing.T) {
	logger, hook := test.NewNullLogger()

	ctx, cancel := context.WithCancel(context.Background())
	wp := WorkerPoolImpl{
		context:       ctx,
		cancelContext: cancel,
		executors:     make(map[string]ExecutorFunc),
		jobChan:       make(chan *Job, 1),
		store:         &dummyStore{},
	}

	WithJobStatusWebhook("http://localhost", time.Minute)(&wp)
	WithLogger(logger)(&wp)

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

	if err := wp.process(job); err != nil {
		t.Fatal(err)
	}

	if len(wp.jobChan) == 0 {
		t.Fatal("expected job channel to contain a job")
	}

	sendNotificationJob := <-wp.jobChan

	if sendNotificationJob.Type != "send_job_status" {
		t.Fatalf("expected pool to have a send_job_status job")
	}

	if err := wp.process(sendNotificationJob); err != nil {
		t.Fatal(err)
	}

	if !sendNotificationCalled {
		t.Fatalf("expected 'sendNotificationCalled' to equal true")
	}

	if len(hook.Entries) > 0 {
		t.Fatalf("did not expect a warning, got %s", hook.LastEntry().Message)
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

		logger, hook := test.NewNullLogger()

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPoolImpl{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)
		WithLogger(logger)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return nil
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		if err := wp.process(job); err != nil {
			t.Fatal(err)
		}

		if err := wp.process(<-wp.jobChan); err != nil {
			t.Fatal(err)
		}

		if webhookJob.Type != "TestJobType" {
			t.Fatalf("expected webhook endpoint to have received a notification")
		}

		if webhookJob.State != Complete {
			t.Fatalf("expected job to be in state '%s' got '%s'", Complete, webhookJob.State)
		}

		if len(hook.Entries) > 0 {
			t.Fatalf("did not expect a warning, got %s", hook.LastEntry().Message)
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

		logger, hook := test.NewNullLogger()

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPoolImpl{
			context:       ctx,
			cancelContext: cancel,
			executors:     make(map[string]ExecutorFunc),
			jobChan:       make(chan *Job, 1),
			store:         &dummyStore{},
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)
		WithLogger(logger)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return ErrPermanentFailure
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		if err := wp.process(job); err != nil {
			t.Fatal(err)
		}
		if err := wp.process(<-wp.jobChan); err != nil {
			t.Fatal(err)
		}

		if webhookJob.Type != "TestJobType" {
			t.Fatalf("expected webhook endpoint to have received a notification")
		}

		if webhookJob.State != Failed {
			t.Fatalf("expected job to be in state '%s' got '%s'", Failed, webhookJob.State)
		}

		if len(hook.Entries) == 0 {
			t.Fatalf("expected a warning")
		}
	})

	t.Run("erroring job should not send", func(t *testing.T) {
		logger, _ := test.NewNullLogger()

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPoolImpl{
			context:          ctx,
			cancelContext:    cancel,
			executors:        make(map[string]ExecutorFunc),
			jobChan:          make(chan *Job, 1),
			store:            &dummyStore{},
			maxJobErrorCount: 1,
		}

		WithJobStatusWebhook("http://localhost", time.Minute)(&wp)
		WithLogger(logger)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return fmt.Errorf("test error")
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		if err := wp.process(job); err != nil {
			t.Fatal(err)
		}

		if len(wp.jobChan) != 0 {
			t.Errorf("did not expect a job to be queued")
		}
	})

	t.Run("valid job should send but get an endpoint error and retry send", func(t *testing.T) {
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "test error", http.StatusBadGateway)
		}))
		defer svr.Close()

		logger, hook := test.NewNullLogger()

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPoolImpl{
			context:          ctx,
			cancelContext:    cancel,
			executors:        make(map[string]ExecutorFunc),
			jobChan:          make(chan *Job, 1),
			store:            &dummyStore{},
			maxJobErrorCount: 1,
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)
		WithLogger(logger)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true
			return nil
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		if err := wp.process(job); err != nil {
			t.Fatal()
		}

		sendNotificationJob := <-wp.jobChan

		if err := wp.process(sendNotificationJob); err != nil {
			t.Fatal(err)
		}

		if len(hook.Entries) != 1 {
			t.Errorf("expected there to be one warning, got %d", len(hook.Entries))
		}

		if sendNotificationJob.State != Error {
			t.Errorf("expected send notification job to be in '%s' state, got '%s'", Error, sendNotificationJob.State)
		}
	})
}

func TestJobErrorMessages(t *testing.T) {
	t.Run("all error messages are stored & published when retries occur", func(t *testing.T) {
		retryCount := 3
		logger, hook := test.NewNullLogger()
		expectedErrorMessages := []string{}

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var resp JSONResponse
			err := json.NewDecoder(r.Body).Decode(&resp)
			if err != nil {
				panic(err)
			}

			if !reflect.DeepEqual(resp.Errors, expectedErrorMessages) {
				t.Errorf("error messages don't match the expected error messages, expected: %v, got: %v", expectedErrorMessages, resp.Errors)
			}
		}))
		defer svr.Close()

		ctx, cancel := context.WithCancel(context.Background())
		wp := WorkerPoolImpl{
			context:          ctx,
			cancelContext:    cancel,
			executors:        make(map[string]ExecutorFunc),
			jobChan:          make(chan *Job, 1),
			store:            &dummyStore{},
			maxJobErrorCount: retryCount,
		}

		WithJobStatusWebhook(svr.URL, time.Minute)(&wp)
		WithLogger(logger)(&wp)

		wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

		wp.RegisterExecutor("TestJobType", func(ctx context.Context, j *Job) error {
			j.ShouldSendNotification = true

			// Fail the first n times, n = retryCount
			if j.ExecCount <= retryCount {
				errorMessage := fmt.Sprintf("error message %d", j.ExecCount)
				expectedErrorMessages = append(expectedErrorMessages, errorMessage)
				return fmt.Errorf(errorMessage)
			}

			j.Result = "done"

			return nil
		})

		job, err := wp.CreateJob("TestJobType", "")
		if err != nil {
			t.Fatal(err)
		}

		// Explicitly retry to trigger n errors and a final successful execution, n = retryCount
		for n := 0; n < retryCount+1; n++ {
			if err := wp.process(job); err != nil {
				t.Fatal(err)
			}
		}

		// Send the notification
		sendNotificationJob := <-wp.jobChan
		if err := wp.process(sendNotificationJob); err != nil {
			t.Fatal(err)
		}

		// Check log entries
		if len(hook.Entries) != retryCount {
			t.Errorf("expected there to be %d warning(s), got %d", retryCount, len(hook.Entries))
		}

		// Final value for "Error" should be blank
		if job.Error != "" {
			t.Errorf("expected job.Error to be blank, got: %#v", job.Error)
		}

		// Check stored error messages
		errorMessages := []string(job.Errors)
		if !reflect.DeepEqual([]string(job.Errors), expectedErrorMessages) {
			t.Errorf("error messages don't match the expected error messages, expected: %v, got: %v", expectedErrorMessages, errorMessages)
		}
	})
}
