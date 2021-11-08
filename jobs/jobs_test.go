package jobs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	wp := WorkerPool{
		executors: make(map[string]ExecutorFunc),
		jobChan:   make(chan *Job, 1),
		logger:    log.New(writer, "", 0),
		store:     &dummyStore{},
	}

	WithJobStatusWebhook("http://localhost")(&wp)

	sendNotificationCalled := false

	wp.RegisterExecutor(SendJobStatusJobType, func(j *Job) error {
		j.ShouldSendNotification = false
		sendNotificationCalled = true
		return nil
	})

	wp.RegisterExecutor("TestJobType", func(j *Job) error {
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
	notification := ""

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		notification = string(bodyBytes)
		fmt.Fprintf(w, "ok")
	}))
	defer svr.Close()

	writer := &dummyWriter{T: t}

	wp := WorkerPool{
		executors: make(map[string]ExecutorFunc),
		jobChan:   make(chan *Job, 1),
		logger:    log.New(writer, "", 0),
		store:     &dummyStore{},
	}

	WithJobStatusWebhook(svr.URL)(&wp)

	wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

	wp.RegisterExecutor("TestJobType", func(j *Job) error {
		j.ShouldSendNotification = true
		return nil
	})

	job, err := wp.CreateJob("TestJobType", "")
	if err != nil {
		t.Fatal(err)
	}

	wp.process(job)
	wp.process(<-wp.jobChan)

	if notification == "" || !strings.Contains(notification, "TestJobType") {
		t.Fatalf("expected webhook endpoint to have received a notification")
	}

	if len(writer.record) > 0 {
		t.Fatalf("did not expect a warning, got %s", writer.record)
	}
}

func TestExecuteSendNotificationFail(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test error", http.StatusBadGateway)
	}))
	defer svr.Close()

	writer := &dummyWriter{T: t}

	wp := WorkerPool{
		executors: make(map[string]ExecutorFunc),
		jobChan:   make(chan *Job, 1),
		logger:    log.New(writer, "", 0),
		store:     &dummyStore{},
	}

	WithJobStatusWebhook(svr.URL)(&wp)

	wp.RegisterExecutor(SendJobStatusJobType, wp.executeSendJobStatus)

	wp.RegisterExecutor("TestJobType", func(j *Job) error {
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
}
