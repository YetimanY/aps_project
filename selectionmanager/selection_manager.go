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
		loop:
			for i := 0; ; i = (i + 1) % cap(pm.processingSystemPool) {
				select {
				case pm.processingSystemPool[i].RequestPort <- req:
					pm.logger <- logger.LogInfo{AuthorID: uint64(i + 1), Req: req, CurrentTime: time.Now(), OperationType: logger.Processing}
					break loop
				default:
					continue loop
				}
			}
		}

		for i := 0; i < cap(pm.processingSystemPool); i++ {
			close(pm.processingSystemPool[i].RequestPort)
			for pm.processingSystemPool[i].Alive == true {
			}
		}
		pm.wg.Done()
	}()
	return true
}
