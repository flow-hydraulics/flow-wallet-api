package ops

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type OpsJob func() error

type OpsWorkerPoolService interface {
	Start()
	Stop()
	AddJob(job OpsJob)
}

type workerPoolImpl struct {
	numWorkers uint

	jobChan chan OpsJob
	wg      *sync.WaitGroup
}

func NewWorkerPool(
	numWorkers uint,
) OpsWorkerPoolService {
	return &workerPoolImpl{
		numWorkers: numWorkers,
		jobChan:    make(chan OpsJob),
		wg:         &sync.WaitGroup{},
	}
}

func (p *workerPoolImpl) Start() {
	for i := uint(0); i < p.numWorkers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobChan {
				if job == nil {
					break
				}

				if err := job(); err != nil {
					log.Warnf("Error running ops job: %s", err)
				}
			}
		}()
	}
}

func (p *workerPoolImpl) Stop() {
	close(p.jobChan)
	p.wg.Wait()
}

func (p *workerPoolImpl) AddJob(job OpsJob) {
	p.jobChan <- job
}
