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
}

type workerPoolImpl struct {
	numWorkers uint

	fungibleInitJobChan chan OpsInitFungibleVaultsJob
	wg                  *sync.WaitGroup
}

func NewWorkerPool(
	numWorkers uint,
) OpsWorkerPoolService {
	return &workerPoolImpl{
		numWorkers:          numWorkers,
		fungibleInitJobChan: make(chan OpsInitFungibleVaultsJob),
		wg:                  &sync.WaitGroup{},
	}
}

func (p *workerPoolImpl) Start() {
	for i := uint(0); i < p.numWorkers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
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
	p.wg.Wait()
}

func (p *workerPoolImpl) AddFungibleInitJob(job OpsInitFungibleVaultsJob) {
	p.fungibleInitJobChan <- job
}
