package productionmanager

import (
	"aps/logger"
	"aps/queue"
	"aps/request"
	"sync"
)

type ProductionManager struct {
	port       chan request.Request
	buffer     *queue.CyclicQueue
	logger     logger.Logger
	numRequest uint64
	wg         *sync.WaitGroup
}

func CreatePM(port chan request.Request, buf *queue.CyclicQueue, logger logger.Logger, wg *sync.WaitGroup) *ProductionManager {
	return &ProductionManager{port: port, buffer: buf, logger: logger, wg: wg}
}

func (pm *ProductionManager) Start() bool {
	go func() {
		for {
			req, ok := <-pm.port
			if !ok {
				break
			}
			pm.numRequest++
			pm.buffer.AddRequest(req)
		}
		pm.wg.Done()
	}()
	return true
}

func (pm *ProductionManager) NumRequest() uint64 {
	return pm.numRequest
}
