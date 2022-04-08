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
	wallet_errors "github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/system"
)

var (
	ErrInvalidJobType   = errors.New("invalid job type")
	ErrPermanentFailure = errors.New("permanent failure")

	// maxJobErrorCount is the maximum number of times a Job can be tried to
	// execute before considering it completely failed.
	defaultMaxJobErrorCount = 10

	// Poll DB for new schedulable jobs every 30s.
	defaultDBJobPollInterval = 30 * time.Second

	// Grace time period before re-scheduling jobs that are in state INIT or
	// ACCEPTED. These are jobs where the executor processing has been
	// unexpectedly disrupted (such as bug, dead node, disconnected
	// networking etc.).
	defaultAcceptedGracePeriod = 3 * time.Minute

	// Grace time period before re-scheduling jobs that are up for immediate
	// restart (such as NO_AVAILABLE_WORKERS or ERROR).
	defaultReSchedulableGracePeriod = 1 * time.Minute
)

type ExecutorFunc func(ctx context.Context, j *Job) error

type WorkerPool interface {
	RegisterExecutor(jobType string, executorF ExecutorFunc)
	CreateJob(jobType, txID string, opts ...JobOption) (*Job, error)
	Schedule(j *Job) error
	Status() (WorkerPoolStatus, error)
	Start()
	Stop(wait bool)
	Capacity() uint
	QueueSize() uint
}

type WorkerPoolImpl struct {
	started       bool
	wg            *sync.WaitGroup
	jobChan       chan *Job
	stopChan      chan struct{}
	context       context.Context
	cancelContext context.CancelFunc
	executors     map[string]ExecutorFunc
	logger        *log.Logger

	store       Store
	capacity    uint
	workerCount uint

	maxJobErrorCount         int
	dbJobPollInterval        time.Duration
	acceptedGracePeriod      time.Duration
	reSchedulableGracePeriod time.Duration

	notificationConfig *NotificationConfig
	systemService      system.Service
}

type WorkerPoolStatus struct {
	JobQueueStatus
	Capacity    int `json:"poolCapacity"`
	WorkerCount int `json:"workerCount"`
}

func NewWorkerPool(db Store, capacity uint, workerCount uint, opts ...WorkerPoolOption) WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPoolImpl{
		wg:            &sync.WaitGroup{},
		jobChan:       make(chan *Job, capacity),
		stopChan:      make(chan struct{}),
		context:       ctx,
		cancelContext: cancel,
		executors:     make(map[string]ExecutorFunc),
		logger:        log.StandardLogger(),

		store:       db,
		capacity:    capacity,
		workerCount: workerCount,

		maxJobErrorCount:         defaultMaxJobErrorCount,
		dbJobPollInterval:        defaultDBJobPollInterval,
		acceptedGracePeriod:      defaultAcceptedGracePeriod,
		reSchedulableGracePeriod: defaultReSchedulableGracePeriod,

		notificationConfig: &NotificationConfig{},
	}

	// Go through options
	for _, opt := range opts {
		opt(pool)
	}

	// Register asynchronous job executor.
	pool.RegisterExecutor(SendJobStatusJobType, pool.executeSendJobStatus)

	pool.logger.Debug(pool)

	return pool
}

func (wp *WorkerPoolImpl) Status() (WorkerPoolStatus, error) {
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
func (wp *WorkerPoolImpl) CreateJob(jobType, txID string, opts ...JobOption) (*Job, error) {
	// Init job
	job := &Job{
		State:         Init,
		Type:          jobType,
		TransactionID: txID,
	}

	// Go through options
	for _, opt := range opts {
		opt(job)
	}

	// Insert job into database
	if err := wp.store.InsertJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

func (wp *WorkerPoolImpl) RegisterExecutor(jobType string, executorF ExecutorFunc) {
	wp.executors[jobType] = executorF
}

// Schedule will try to immediately schedule the run of a job
func (wp *WorkerPoolImpl) Schedule(j *Job) error {
	entry := j.logEntry(wp.logger.WithFields(log.Fields{
		"package":  "jobs",
		"function": "WorkerPool.Schedule",
	}))

	entry.Debug("Scheduling job")

	if halted, err := wp.systemHalted(); err != nil {
		return fmt.Errorf("error while getting system settings: %w", err)
	} else if halted {
		// System halted; prevent immediate scheduling and let dbScheduler handle this job
		entry.Debug("System halted")
		return nil
	}

	if !wp.tryEnqueue(j, false) {
		j.State = NoAvailableWorkers
		entry.Debug("No available workers, deferring")
		if err := wp.store.UpdateJob(j); err != nil {
			return err
		}
	} else {
		entry.Debug("Successfully scheduled job")
	}

	return nil
}

func (wp *WorkerPoolImpl) Start() {
	if !wp.started {
		wp.started = true
		wp.startWorkers()
		wp.startDBJobScheduler()
	}
}

func (wp *WorkerPoolImpl) Stop(wait bool) {
	close(wp.stopChan)
	// Give time for the stop channel to signal before closing job channel
	time.Sleep(time.Millisecond * 100)
	close(wp.jobChan)
	if wait {
		wp.cancelContext()
		wp.wg.Wait()
	}
}

func (wp *WorkerPoolImpl) Capacity() uint {
	return wp.capacity
}

func (wp *WorkerPoolImpl) QueueSize() uint {
	return uint(len(wp.jobChan))
}

func (wp *WorkerPoolImpl) accept(job *Job) bool {
	entry := job.logEntry(wp.logger.WithFields(log.Fields{
		"package":  "jobs",
		"function": "WorkerPool.accept",
	}))

	if err := wp.store.AcceptJob(job, wp.acceptedGracePeriod); err != nil {
		entry.
			WithFields(log.Fields{"error": err}).
			Warn("Failed to accept job")
		return false
	}

	return true
}

func (wp *WorkerPoolImpl) systemHalted() (bool, error) {
	if wp.systemService != nil {
		return wp.systemService.IsHalted()
	}
	return false, nil
}

func (wp *WorkerPoolImpl) startDBJobScheduler() {
	go func() {
		var restTime time.Duration

	jobPoolLoop:
		for {
			select {
			case <-time.After(restTime):
			case <-wp.stopChan:
				break jobPoolLoop
			}

			if halted, err := wp.systemHalted(); err != nil {
				wp.logger.
					WithFields(log.Fields{"error": err}).
					Warn("Could not get system settings from DB")
				restTime = wp.dbJobPollInterval
				continue
			} else if halted {
				restTime = wp.dbJobPollInterval
				continue
			}

			begin := time.Now()

			o := datastore.ParseListOptions(0, 0)
			jobs, err := wp.store.SchedulableJobs(wp.acceptedGracePeriod, wp.reSchedulableGracePeriod, o)
			if err != nil {
				wp.logger.
					WithFields(log.Fields{"error": err}).
					Warn("Could not fetch schedulable jobs from DB")
				continue
			}

			for i := range jobs {
				wp.tryEnqueue(&jobs[i], true)
			}

			elapsed := time.Since(begin)
			restTime = wp.dbJobPollInterval - elapsed
		}
	}()
}

func (wp *WorkerPoolImpl) startWorkers() {
	for i := uint(0); i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for job := range wp.jobChan {
				if job == nil {
					break
				}

				if err := wp.process(job); err != nil {
					// Handle critical processing errors

					entry := job.logEntry(wp.logger.WithFields(log.Fields{
						"package":  "jobs",
						"function": "WorkerPool.startWorkers.goroutine",
						"error":    err,
					}))

					if wallet_errors.IsChainConnectionError(err) {
						if wp.systemService != nil {
							entry.Warn("Unable to connect to chain, pausing system")
							entry.Warn(err)
							// Unable to connect to chain, pause system.
							if err := wp.systemService.Pause(); err != nil {
								entry.
									WithFields(log.Fields{"error": err}).
									Warn("Unable to pause system")
							}
						} else {
							entry.Warn("Unable to connect to chain")
						}
					} else {
						entry.Warn("Critical error while processing job")
					}
				}
			}
		}()
	}
}

func (wp *WorkerPoolImpl) tryEnqueue(job *Job, block bool) bool {
	if block {
		select {
		case <-wp.stopChan:
			return false
		case wp.jobChan <- job:
			return true
		}
	} else {
		select {
		case <-wp.stopChan:
			return false
		case wp.jobChan <- job:
			return true
		default:
			return false
		}
	}
}

func (wp *WorkerPoolImpl) process(job *Job) error {
	entry := job.logEntry(wp.logger.WithFields(log.Fields{
		"package":  "jobs",
		"function": "WorkerPool.process",
	}))

	if !wp.accept(job) {
		entry.Info("Failed to accept job")
		return nil
	}

	executor, exists := wp.executors[job.Type]
	if !exists {
		entry.Warn("Could not process job, no registered executor for type")

		job.State = NoAvailableWorkers

		if err := wp.store.UpdateJob(job); err != nil {
			return fmt.Errorf("error while updating database entry: %w", err)
		}

		return nil
	}

	if err := executor(wp.context, job); err != nil {
		// Check for chain connection errors
		if wallet_errors.IsChainConnectionError(err) {
			// Stop processing this job any further, returning it to the pool.
			return err
		}

		if job.ExecCount > wp.maxJobErrorCount || errors.Is(err, ErrPermanentFailure) {
			job.State = Failed
		} else {
			job.State = Error
		}

		job.Error = err.Error()
		job.Errors = append(job.Errors, err.Error())

		entry.
			WithFields(log.Fields{"error": err}).
			Warn("Job execution resulted with error")

	} else {
		job.State = Complete
		job.Error = "" // Clear the error message for the final & successful execution
	}

	if err := wp.store.UpdateJob(job); err != nil {
		return fmt.Errorf("error while updating database entry: %w", err)
	}

	if (job.State == Failed || job.State == Complete) && job.ShouldSendNotification && wp.notificationConfig.ShouldSendJobStatus() {
		if err := wp.scheduleJobStatusNotification(job); err != nil {
			entry.
				WithFields(log.Fields{"error": err}).
				Warn("Could not schedule a status update notification for job")
		}
	}

	return nil
}

func (wp *WorkerPoolImpl) executeSendJobStatus(ctx context.Context, j *Job) error {
	if j.Type != SendJobStatusJobType {
		return ErrInvalidJobType
	}

	j.ShouldSendNotification = false

	return wp.notificationConfig.SendJobStatus(ctx, j.Result)
}

func PermanentFailure(err error) error {
	return fmt.Errorf("%w: %s", ErrPermanentFailure, err.Error())
}

func (wp *WorkerPoolImpl) scheduleJobStatusNotification(parent *Job) error {
	entry := parent.logEntry(wp.logger.WithFields(log.Fields{
		"package":  "jobs",
		"function": "ScheduleJobStatusNotification",
	}))

	entry.Debug("Scheduling job status notification")

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
