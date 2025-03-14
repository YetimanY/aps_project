package selectionmanager

import (
	"aps/logger"
	"aps/processingsystem"
	"aps/queue"
	"sync"
	"sync/atomic"
	"time"
)

type SelectionManager struct {
	processingSystemPool []processingsystem.ProcessingSystem
	buffer               *queue.CyclicQueue
	logger               logger.Logger
	closedFlag           *atomic.Value
	wg                   *sync.WaitGroup
}

func CreateSM(psPool []processingsystem.ProcessingSystem, buf *queue.CyclicQueue, logger logger.Logger, closedFlag *atomic.Value, wg *sync.WaitGroup) SelectionManager {
	return SelectionManager{psPool, buf, logger, closedFlag, wg}
}

func (pm *SelectionManager) Start() bool {
	go func() {
		for pm.closedFlag.Load() == false || !pm.buffer.IsEmpty() {
			req := pm.buffer.GetRequest()
			if req.SourceID() == 0 || req.RequestID() == 0 {
				continue
			}
			for i := 0; ; i = (i + 1) % cap(pm.processingSystemPool) {
				if pm.processingSystemPool[i].Status() == 0 {
					pm.logger <- logger.LogInfo{AuthorID: uint64(i + 1), Req: req, CurrentTime: time.Now(), OperationType: logger.Processing}
					pm.processingSystemPool[i].SendRequest(req)
					break
				}
			}
		}

		for i := 0; i < cap(pm.processingSystemPool); i++ {
			pm.processingSystemPool[i].Interrupt()
			for pm.processingSystemPool[i].Status() != -1 {
			}
		}
		pm.wg.Done()
	}()
	return true
}
