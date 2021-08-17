package jobs

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eqlabs/flow-wallet-api/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerPool struct {
	log         *log.Logger
	wg          *sync.WaitGroup
	store       Store
	jobChan     chan *Job
	capacity    uint
	workerCount uint
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

func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = uuid.New()
	return
}

func (j *Job) Wait(wait bool) error {
	if wait {
		// Wait for the job to have finished
		for j.Status == Accepted {
			time.Sleep(10 * time.Millisecond)
		}
		if j.Status == Error {
			return fmt.Errorf(j.Error)
		}
	}
	return nil
}

func NewWorkerPool(l *log.Logger, db Store, capacity uint, workerCount uint) *WorkerPool {
	wg := &sync.WaitGroup{}
	jobChan := make(chan *Job, capacity)

	pool := &WorkerPool{l, wg, db, jobChan, capacity, workerCount}

	pool.initWorkers()

	return pool
}

func (p *WorkerPool) initWorkers() {
	for i := uint(0); i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.newWorker()
	}
}

func (p *WorkerPool) AddJob(do func() (string, error)) (*Job, error) {
	job := &Job{Do: do, Status: Init}
	if err := p.store.InsertJob(job); err != nil {
		return job, err
	}

	if !p.tryEnqueue(job) {
		job.Status = QueueFull
		if err := p.store.UpdateJob(job); err != nil {
			p.log.Println("WARNING: Could not update DB entry for Job", job.ID)
		}
		return job, &errors.JobQueueFull{Err: fmt.Errorf(job.Status.String())}
	}

	job.Status = Accepted
	if err := p.store.UpdateJob(job); err != nil {
		p.log.Println("WARNING: Could not update DB entry for Job", job.ID)
	}

	return job, nil
}

func (p *WorkerPool) Stop() {
	close(p.jobChan)
	p.wg.Wait()
}

func (p *WorkerPool) newWorker() {
	defer p.wg.Done()
	for job := range p.jobChan {
		if job == nil {
			return
		}
		p.process(job)
	}
}

func (p *WorkerPool) tryEnqueue(job *Job) bool {
	select {
	case p.jobChan <- job:
		return true
	default:
		return false
	}
}

func (p *WorkerPool) process(job *Job) {
	result, err := job.Do()
	job.Result = result
	if err != nil {
		if p.log != nil {
			p.log.Printf("[Job %s] Error while processing job: %s\n", job.ID, err)
		}
		job.Status = Error
		job.Error = err.Error()
		if err := p.store.UpdateJob(job); err != nil {
			p.log.Println("WARNING: Could not update DB entry for Job", job.ID)
		}
		return
	}
	job.Status = Complete
	if err := p.store.UpdateJob(job); err != nil {
		p.log.Println("WARNING: Could not update DB entry for Job", job.ID)
	}
}
