package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func Test_WorkerPoolExecutesJobWithSuccess(t *testing.T) {
	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)
	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)

	executedWG := &sync.WaitGroup{}
	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		defer executedWG.Done()
		return nil
	}

	wp.RegisterExecutor(jobType, jobFunc)

	executedWG.Add(1)
	j, err := wp.CreateJob(jobType, "0xf00d")
	if err != nil {
		t.Fatal(err)
	}

	err = wp.Schedule(j)
	if err != nil {
		t.Fatal(err)
	}

	executedWG.Wait()

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if job.ExecCount > 1 || (time.Since(job.UpdatedAt) < 250*time.Millisecond) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Complete {
		t.Fatalf("expected job.State = %q, got %q", jobs.Complete, job.State)
	}
}

func Test_WorkerPoolExecutesJobWithError(t *testing.T) {
	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)
	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)

	executedWG := &sync.WaitGroup{}
	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		defer executedWG.Done()
		return errors.New("test job executor error returned on purpose")
	}

	wp.RegisterExecutor(jobType, jobFunc)

	executedWG.Add(1)
	j, err := wp.CreateJob(jobType, "0xf00d")
	if err != nil {
		t.Fatal(err)
	}

	err = wp.Schedule(j)
	if err != nil {
		t.Fatal(err)
	}

	executedWG.Wait()

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if job.ExecCount < 1 || (time.Since(job.UpdatedAt) < 250*time.Millisecond) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Error {
		t.Fatalf("expected job.State = %q, got %q", jobs.Error, job.State)
	}
}

func Test_WorkerPoolExecutesJobWithPermanentError(t *testing.T) {
	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)
	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)

	executedWG := &sync.WaitGroup{}
	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		defer executedWG.Done()
		return jobs.PermanentFailure(errors.New("test job executor error returned on purpose"))
	}

	wp.RegisterExecutor(jobType, jobFunc)

	executedWG.Add(1)
	j, err := wp.CreateJob(jobType, "0xf00d")
	if err != nil {
		t.Fatal(err)
	}

	err = wp.Schedule(j)
	if err != nil {
		t.Fatal(err)
	}

	executedWG.Wait()

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if job.ExecCount < 1 || (time.Since(job.UpdatedAt) < 250*time.Millisecond) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Failed {
		t.Fatalf("expected job.State = %q, got %q", jobs.Failed, job.State)
	}
}

func Test_WorkerPoolPicksUpInitJob(t *testing.T) {
	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)

	executedWG := &sync.WaitGroup{}
	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		defer executedWG.Done()
		return nil
	}

	t0 := time.Now()
	j := &jobs.Job{
		ID:            uuid.New(),
		State:         jobs.Init,
		Type:          jobType,
		TransactionID: "0xf00d",
		ExecCount:     2,
		CreatedAt:     t0.Add(-10 * time.Minute),
		UpdatedAt:     t0.Add(-10 * time.Minute),
	}

	// Directly insert "old" job into DB.
	err := db.Create(j).Error
	if err != nil {
		t.Fatal(err)
	}

	executedWG.Add(1)
	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)
	wp.RegisterExecutor(jobType, jobFunc)

	executedWG.Wait()

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if time.Since(job.UpdatedAt) < 250*time.Millisecond {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Complete {
		t.Fatalf("expected job.State = %q, got %q", jobs.Complete, job.State)
	}
}

func Test_WorkerPoolPicksUpErroredJob(t *testing.T) {
	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)

	executedWG := &sync.WaitGroup{}
	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		defer executedWG.Done()
		return nil
	}

	t0 := time.Now()
	j := &jobs.Job{
		ID:            uuid.New(),
		State:         jobs.Error,
		Type:          jobType,
		TransactionID: "0xf00d",
		ExecCount:     2,
		CreatedAt:     t0.Add(-10 * time.Minute),
		UpdatedAt:     t0.Add(-10 * time.Minute),
	}

	// Directly insert "old" job into DB.
	err := db.Create(j).Error
	if err != nil {
		t.Fatal(err)
	}

	executedWG.Add(1)
	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)
	wp.RegisterExecutor(jobType, jobFunc)

	executedWG.Wait()

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if time.Since(job.UpdatedAt) < 250*time.Millisecond {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Complete {
		t.Fatalf("expected job.State = %q, got %q", jobs.Complete, job.State)
	}
}

func Test_WorkerPoolDoesntPickupFailedJob(t *testing.T) {
	// XXX: This test is very much a best effort case. There are several
	// theoretical glitches here that make this a bit unreliable. There's a
	// chance that the whole test executes too quickly when compared to worker
	// pool DB job polling & scheduling.
	//
	// There's also a theoretical chance that the DB job polling and scheduling
	// executes quicker than job executer registration.
	//
	// The last and maybe most difficult and controversial problem is that
	// there's no way to prove that something that was not expected to happen,
	// does not happen, even over longer period of time.

	logger := log.WithFields(log.Fields{
		"version": "test",
	})

	cfg := test.LoadConfig(t, testConfigPath)
	db := test.GetDatabase(t, cfg)
	jobStore := jobs.NewGormStore(db)

	jobType := "job"
	jobFunc := func(ctx context.Context, j *jobs.Job) error {
		t.Fatal("failed job executed")
		return nil
	}

	t0 := time.Now()
	j := &jobs.Job{
		ID:            uuid.New(),
		State:         jobs.Failed,
		Type:          jobType,
		TransactionID: "0xf00d",
		ExecCount:     2,
		CreatedAt:     t0.Add(-10 * time.Minute),
		UpdatedAt:     t0.Add(-10 * time.Minute),
	}

	// Directly insert "old" job into DB.
	err := db.Create(j).Error
	if err != nil {
		t.Fatal(err)
	}

	wp := jobs.NewWorkerPool(logger, jobStore, 10, 10)
	wp.RegisterExecutor(jobType, jobFunc)

	// Gotta give DB job poller a bit of time to catch up.
	time.Sleep(100 * time.Millisecond)

	var job jobs.Job
	for {
		job, err = jobStore.Job(j.ID)
		if err != nil {
			t.Fatal(err)
		}

		if job.ExecCount < 1 || (time.Since(job.UpdatedAt) < 250*time.Millisecond) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if job.State != jobs.Failed {
		t.Fatalf("expected job.State = %q, got %q", jobs.Failed, job.State)
	}
}
