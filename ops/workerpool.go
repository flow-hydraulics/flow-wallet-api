package ops

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type OpsInitFungibleVaultsJobFunc func(address string, tokenList []string) error
type OpsInitFungibleVaultsJob struct {
	Func      OpsInitFungibleVaultsJobFunc
	Address   string
	TokenList []string
}

type OpsWorkerPoolService interface {
	Start()
	Stop()
	AddFungibleInitJob(job OpsInitFungibleVaultsJob)
	NumWorkers() uint
}

type workerPoolImpl struct {
	numWorkers uint
	capacity   uint

	fungibleInitJobChan chan OpsInitFungibleVaultsJob
	workersWaitGroup    *sync.WaitGroup
}

func NewWorkerPool(
	numWorkers uint,
	capacity uint,
) OpsWorkerPoolService {
	return &workerPoolImpl{
		numWorkers:          numWorkers,
		capacity:            capacity,
		fungibleInitJobChan: make(chan OpsInitFungibleVaultsJob, capacity),
		workersWaitGroup:    &sync.WaitGroup{},
	}
}

func (p *workerPoolImpl) Start() {
	for i := uint(0); i < p.numWorkers; i++ {
		p.workersWaitGroup.Add(1)
		go func() {
			defer p.workersWaitGroup.Done()
			for job := range p.fungibleInitJobChan {
				if job.Func == nil {
					break
				}

				if err := job.Func(job.Address, job.TokenList); err != nil {
					log.Warnf("Error running ops job: %s", err)
				}
			}
		}()
	}
}

func (p *workerPoolImpl) Stop() {
	close(p.fungibleInitJobChan)
	p.workersWaitGroup.Wait()
}

func (p *workerPoolImpl) AddFungibleInitJob(job OpsInitFungibleVaultsJob) {
	p.fungibleInitJobChan <- job
}

func (p *workerPoolImpl) NumWorkers() uint {
	return p.numWorkers
}
