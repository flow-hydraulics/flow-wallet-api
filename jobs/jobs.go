package jobs

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	STATUS_INIT                 = "init"
	STATUS_ACCEPTED             = "accepted"
	STATUS_NO_AVAILABLE_WORKERS = "no-available-workers"
	STATUS_QUEUE_FULL           = "queue-full"
	STATUS_ERROR                = "error"
	STATUS_COMPLETE             = "complete"
)

type WorkerPool struct {
	wg      *sync.WaitGroup
	workers []*Worker
	log     *log.Logger
	db      JobStore
}

type Worker struct {
	pool    *WorkerPool
	jobChan *chan *Job
}

type Job struct {
	ID        uuid.UUID              `json:"jobId" gorm:"primary_key;type:uuid;"`
	Do        func() (string, error) `json:"-" gorm:"-"`
	Status    string                 `json:"status"`
	Error     string                 `json:"-"`
	Result    string                 `json:"result"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	DeletedAt gorm.DeletedAt         `json:"-" gorm:"index"`
}

func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = uuid.New()
	return
}

func NewWorkerPool(log *log.Logger, db JobStore) *WorkerPool {
	return &WorkerPool{&sync.WaitGroup{}, []*Worker{}, log, db}
}

func (p *WorkerPool) AddWorker(capacity uint) {
	if len(p.workers) > 0 {
		panic("multiple workers not supported yet")
	}
	p.wg.Add(1)
	jobChan := make(chan *Job, capacity)
	worker := &Worker{p, &jobChan}
	p.workers = append(p.workers, worker)
	go worker.start()
}

func (p *WorkerPool) AddJob(do func() (string, error)) (*Job, error) {
	job := &Job{Do: do, Status: STATUS_INIT}
	p.db.InsertJob(job)
	worker, err := p.AvailableWorker()
	if err != nil {
		p.log.Println(err)
		job.Status = STATUS_NO_AVAILABLE_WORKERS
		p.db.UpdateJob(job)
		return job, &errors.JobQueueFull{Err: fmt.Errorf(job.Status)}
	}
	if !worker.tryEnqueue(job) {
		job.Status = STATUS_QUEUE_FULL
		p.db.UpdateJob(job)
		return job, &errors.JobQueueFull{Err: fmt.Errorf(job.Status)}
	}
	job.Status = STATUS_ACCEPTED
	p.db.UpdateJob(job)
	return job, nil
}

func (p *WorkerPool) AvailableWorker() (*Worker, error) {
	// TODO: support multiple workers
	if len(p.workers) < 1 {
		return nil, fmt.Errorf("no available workers")
	}
	return p.workers[0], nil
}

func (p *WorkerPool) Stop() {
	for _, w := range p.workers {
		close(*w.jobChan)
	}
	p.wg.Wait()
}

func (w *Worker) start() {
	defer w.pool.wg.Done()
	for job := range *w.jobChan {
		if job == nil {
			return
		}
		w.process(job)
	}
}

func (w *Worker) tryEnqueue(job *Job) bool {
	select {
	case *w.jobChan <- job:
		return true
	default:
		return false
	}
}

func (w *Worker) process(job *Job) {
	result, err := job.Do()
	if err != nil {
		job.Status = STATUS_ERROR
		job.Error = err.Error()
		w.pool.db.UpdateJob(job)
		return
	}
	job.Status = STATUS_COMPLETE
	job.Result = result
	w.pool.db.UpdateJob(job)
}
