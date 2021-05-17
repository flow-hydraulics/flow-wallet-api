package jobs

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/eqlabs/flow-wallet-service/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerPool struct {
	wg      *sync.WaitGroup
	workers []*Worker
	db      Store
}

type Worker struct {
	pool    *WorkerPool
	jobChan *chan *Job
}

type Job struct {
	ID        uuid.UUID              `json:"jobId" gorm:"primary_key;type:uuid;"`
	Do        func() (string, error) `json:"-" gorm:"-"`
	Status    Status                 `json:"status"`
	Error     string                 `json:"-"`
	Result    string                 `json:"result"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	DeletedAt gorm.DeletedAt         `json:"-" gorm:"index"`
}

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[JOBS] ", log.LstdFlags|log.Lshortfile)
}

func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = uuid.New()
	return
}

func NewWorkerPool(db Store) *WorkerPool {
	return &WorkerPool{&sync.WaitGroup{}, []*Worker{}, db}
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
	job := &Job{Do: do, Status: Init}
	p.db.InsertJob(job)
	worker, err := p.AvailableWorker()
	if err != nil {
		job.Status = NoAvailableWorkers
		p.db.UpdateJob(job)
		return job, &errors.JobQueueFull{Err: fmt.Errorf(job.Status.String())}
	}
	if !worker.tryEnqueue(job) {
		job.Status = QueueFull
		p.db.UpdateJob(job)
		return job, &errors.JobQueueFull{Err: fmt.Errorf(job.Status.String())}
	}
	job.Status = Accepted
	p.db.UpdateJob(job)
	return job, nil
}

func (p *WorkerPool) AvailableWorker() (*Worker, error) {
	// TODO: support multiple workers, use load balancing
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
		logger.Printf("[Job %s] Error while processing job: %s", job.ID, err)
		job.Status = Error
		job.Error = err.Error()
		w.pool.db.UpdateJob(job)
		return
	}
	job.Status = Complete
	job.Result = result
	w.pool.db.UpdateJob(job)
}
