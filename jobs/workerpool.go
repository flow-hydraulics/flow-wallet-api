package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/system"
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
	acceptedGracePeriod = 3 * time.Minute

	// Grace time period before re-scheduling jobs that are up for immediate
	// restart (such as NO_AVAILABLE_WORKERS or ERROR).
	reSchedulableGracePeriod = 1 * time.Minute
)

type ExecutorFunc func(ctx context.Context, j *Job) error

type WorkerPool struct {
	wg            *sync.WaitGroup
	jobChan       chan *Job
	stopChan      chan struct{}
	context       context.Context
	cancelContext context.CancelFunc
	executors     map[string]ExecutorFunc
	logger        *log.Logger

	store             Store
	capacity          uint
	workerCount       uint
	dbJobPollInterval time.Duration

	notificationConfig *NotificationConfig
	systemService      *system.Service
}

type WorkerPoolStatus struct {
	JobQueueStatus
	Capacity    int `json:"poolCapacity"`
	WorkerCount int `json:"workerCount"`
}

func NewWorkerPool(db Store, capacity uint, workerCount uint, opts ...WorkerPoolOption) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		wg:            &sync.WaitGroup{},
		jobChan:       make(chan *Job, capacity),
		stopChan:      make(chan struct{}),
		context:       ctx,
		cancelContext: cancel,
		executors:     make(map[string]ExecutorFunc),
		logger:        nil,

		store:             db,
		capacity:          capacity,
		workerCount:       workerCount,
		dbJobPollInterval: defaultDBJobPollInterval,

		notificationConfig: &NotificationConfig{},
	}

	pool.startWorkers()
	pool.startDBJobScheduler()

	// Register asynchronous job executor.
	pool.RegisterExecutor(SendJobStatusJobType, pool.executeSendJobStatus)

	// Go through options
	for _, opt := range opts {
		opt(pool)
	}

	if pool.logger == nil {
		pool.logger = log.New()
	}

	return pool
}

func (wp *WorkerPool) Status() (WorkerPoolStatus, error) {
	var status WorkerPoolStatus

	query, err := wp.store.Status()
	if err != nil {
		return status, err
	}

	for _, r := range query {
		switch r.State {
		case Init:
			status.JobsInit = r.Count
		case NoAvailableWorkers:
			status.JobsNotAccepted = r.Count
		case Accepted:
			status.JobsAccepted = r.Count
		case Error:
			status.JobsErrored = r.Count
		case Failed:
			status.JobsFailed = r.Count
		case Complete:
			status.JobsCompleted = r.Count
		default:
			continue
		}
	}

	status.Capacity = int(wp.capacity)
	status.WorkerCount = int(wp.workerCount)

	return status, nil
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

// Schedule will try to immediately schedule the run of a job
func (wp *WorkerPool) Schedule(j *Job) error {
	if wp.waitMaintenance() {
		// In maintenance mode; prevent immediate scheduling and let dbScheduler handle this job
		return nil
	}

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
	wp.cancelContext()
	wp.wg.Wait()
}

func (wp *WorkerPool) Capacity() uint {
	return wp.capacity
}

func (wp *WorkerPool) QueueSize() uint {
	return uint(len(wp.jobChan))
}

func (wp *WorkerPool) accept(job *Job) bool {
	err := wp.store.IncreaseExecCount(job)
	if err != nil {
		wp.logger.
			WithFields(log.Fields{"error": err, "jobID": job.ID}).
			Warn("Failed to increase job exec_count")
		return false
	}

	return true
}

func (wp *WorkerPool) waitMaintenance() bool {
	return wp.systemService != nil && wp.systemService.IsMaintenanceMode()
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

			// Check for maintenance mode
			if wp.waitMaintenance() {
				restTime = wp.dbJobPollInterval
				continue
			}

			begin := time.Now()

			o := datastore.ParseListOptions(0, 0)
			jobs, err := wp.store.SchedulableJobs(acceptedGracePeriod, reSchedulableGracePeriod, o)
			if err != nil {
				wp.logger.
					WithFields(log.Fields{"error": err}).
					Warn("Could not fetch schedulable jobs from DB")
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
		wp.logger.
			WithFields(log.Fields{"jobID": job.ID, "jobType": job.Type}).
			Info("Failed to accept job")
		return
	}

	executor, exists := wp.executors[job.Type]
	if !exists {
		wp.logger.
			WithFields(log.Fields{"jobID": job.ID, "jobType": job.Type}).
			Warn("Could not process job, no registered executor for type")
		job.State = NoAvailableWorkers
		if err := wp.store.UpdateJob(job); err != nil {
			wp.logger.
				WithFields(log.Fields{"error": err, "jobID": job.ID}).
				Warn("Could not update DB entry for job")
		}
		return
	}

	err := executor(wp.context, job)
	if err != nil {
		if job.ExecCount > maxJobErrorCount || errors.Is(err, ErrPermanentFailure) {
			job.State = Failed
		} else {
			job.State = Error
		}
		job.Error = err.Error()
		wp.logger.
			WithFields(log.Fields{"error": err, "jobID": job.ID, "jobType": job.Type}).
			Warn("Job execution resulted with error")
	} else {
		job.State = Complete
		job.Error = ""
	}

	if err := wp.store.UpdateJob(job); err != nil {
		wp.logger.
			WithFields(log.Fields{"error": err, "jobID": job.ID, "jobType": job.Type}).
			Warn("Could not update DB entry for job")
	}

	if (job.State == Failed || job.State == Complete) && job.ShouldSendNotification && wp.notificationConfig.ShouldSendJobStatus() {
		if err := ScheduleJobStatusNotification(wp, job); err != nil {
			wp.logger.
				WithFields(log.Fields{"error": err, "jobID": job.ID, "jobType": job.Type}).
				Warn("Could not schedule a status update notification for job")
		}
	}
}

func (wp *WorkerPool) executeSendJobStatus(ctx context.Context, j *Job) error {
	if j.Type != SendJobStatusJobType {
		return ErrInvalidJobType
	}

	j.ShouldSendNotification = false

	return wp.notificationConfig.SendJobStatus(ctx, j.Result)
}

func PermanentFailure(err error) error {
	return fmt.Errorf("%w: %s", ErrPermanentFailure, err.Error())
}

func ScheduleJobStatusNotification(wp *WorkerPool, parent *Job) error {
	job, err := wp.CreateJob(SendJobStatusJobType, "")
	if err != nil {
		return err
	}

	b, err := json.Marshal(parent.ToJSONResponse())
	if err != nil {
		return err
	}

	// Store the notification content of the parent job in Result of the new job
	job.Result = string(b)

	if err := wp.store.UpdateJob(job); err != nil {
		return err
	}

	return wp.Schedule(job)
}
