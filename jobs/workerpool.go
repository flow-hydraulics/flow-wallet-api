package jobs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
)

// maxJobErrorCount is the maximum number of times a Job can be tried to
// execute before considering it completely failed.
const maxJobErrorCount = 10

var (
	ErrInvalidJobType   = errors.New("invalid job type")
	ErrPermanentFailure = errors.New("permanent failure")

	// Poll DB for new schedulable jobs every 30s.
	defaultDBJobPollInterval = 30 * time.Second

	// Grace time period before re-scheduling jobs that are in state INIT or
	// ACCEPTED. These are jobs where the executor processing has been
	// unexpectedly disrupted (such as bug, dead node, disconnected
	// networking etc.).
	acceptedGracePeriod = 10 * time.Minute

	// Grace time period before re-scheduling jobs that are up for immediate
	// restart (such as NO_AVAILABLE_WORKERS or ERROR).
	reSchedulableGracePeriod = 1 * time.Minute
)

type ExecutorFunc func(j *Job) error

type WorkerPool struct {
	executors map[string]ExecutorFunc

	logger            *log.Logger
	wg                *sync.WaitGroup
	store             Store
	jobChan           chan *Job
	capacity          uint
	workerCount       uint
	dbJobPollInterval time.Duration
	stopChan          chan struct{}
}

func NewWorkerPool(logger *log.Logger, db Store, capacity uint, workerCount uint) *WorkerPool {
	if logger == nil {
		// Make sure we always have a logger
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}
	wg := &sync.WaitGroup{}
	jobChan := make(chan *Job, capacity)
	pool := &WorkerPool{
		executors: make(map[string]ExecutorFunc),

		logger:            logger,
		wg:                wg,
		store:             db,
		jobChan:           jobChan,
		capacity:          capacity,
		workerCount:       workerCount,
		dbJobPollInterval: defaultDBJobPollInterval,
		stopChan:          make(chan struct{}),
	}

	pool.startWorkers()
	pool.startDBJobScheduler()
	return pool
}

// CreateJob constructs a new Job for type `jobType` ready for scheduling.
func (wp *WorkerPool) CreateJob(jobType, txID string) (*Job, error) {
	// Init job
	job := &Job{
		State:         Init,
		Type:          jobType,
		TransactionID: txID,
	}

	// Insert job into database
	if err := wp.store.InsertJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

func (wp *WorkerPool) RegisterExecutor(jobType string, executorF ExecutorFunc) {
	wp.executors[jobType] = executorF
}

func (wp *WorkerPool) Schedule(j *Job) error {
	if !wp.tryEnqueue(j, false) {
		j.State = NoAvailableWorkers
		if err := wp.store.UpdateJob(j); err != nil {
			return err
		}
	}

	return nil
}

func (wp *WorkerPool) Stop() {
	close(wp.stopChan)
	close(wp.jobChan)
	wp.wg.Wait()
}

func (wp *WorkerPool) accept(job *Job) bool {
	err := wp.store.IncreaseExecCount(job)
	if err != nil {
		wp.logger.Printf("WARNING: Failed to increase Job %q exec_count: %s\n", job.ID, err.Error())
		return false
	}

	return true
}

func (wp *WorkerPool) startDBJobScheduler() {
	go func() {
		var restTime time.Duration

	jobPoolLoop:
		for {
			select {
			case <-time.After(restTime):
			case <-wp.stopChan:
				break jobPoolLoop
			}

			begin := time.Now()

			o := datastore.ParseListOptions(0, 0)
			jobs, err := wp.store.SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod, o)
			if err != nil {
				wp.logger.Println("WARNING: Could not fetch schedulable jobs from DB: ", err)
				continue
			}

			for _, j := range jobs {
				wp.tryEnqueue(&j, true)
			}

			elapsed := time.Since(begin)
			restTime = wp.dbJobPollInterval - elapsed
		}
	}()
}

func (wp *WorkerPool) startWorkers() {
	for i := uint(0); i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for job := range wp.jobChan {
				if job == nil {
					break
				}

				wp.process(job)
			}
		}()
	}
}

func (wp *WorkerPool) tryEnqueue(job *Job, block bool) bool {
	if block {
		wp.jobChan <- job
	} else {
		select {
		case wp.jobChan <- job:
			return true
		default:
			return false
		}
	}

	return true
}

func (wp *WorkerPool) process(job *Job) {
	if !wp.accept(job) {
		wp.logger.Printf("INFO: Failed to accept Job(id: %q, type: %q)", job.ID, job.Type)
		return
	}

	executor, exists := wp.executors[job.Type]
	if !exists {
		wp.logger.Printf("WARNING: Could not process Job %q: no registered executor for type %q", job.ID, job.Type)
		job.State = NoAvailableWorkers
		if err := wp.store.UpdateJob(job); err != nil {
			wp.logger.Printf("WARNING: Could not update DB entry for Job %q: %s\n", job.ID, err.Error())
		}
		return
	}

	err := executor(job)
	if err != nil {
		if job.ExecCount > maxJobErrorCount || errors.Is(err, ErrPermanentFailure) {
			job.State = Failed
		} else {
			job.State = Error
		}
		job.Error = err.Error()
		wp.logger.Printf("WARNING: Job(id: %q, type: %q) execution resulted with error: %s", job.ID, job.Type, job.Error)
	} else {
		job.State = Complete
		job.Error = ""
	}

	if err := wp.store.UpdateJob(job); err != nil {
		wp.logger.Printf("WARNING: Could not update DB entry for Job(id: %q, type: %q): %s\n", job.ID, job.Type, err.Error())
	}
}

func PermanentFailure(err error) error {
	return fmt.Errorf("%w: %s", ErrPermanentFailure, err.Error())
}
