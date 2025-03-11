package processingsystem

import (
	"aps/logger"
	"aps/request"
	"time"
)

type ProcessingSystem struct {
	ID          uint64
	RequestPort chan request.Request
	LoggerPort  logger.Logger
	Alive       bool
}

func CreateProcessingSystem(id uint64, requestPort chan request.Request, loggerPort logger.Logger) *ProcessingSystem {
	return &ProcessingSystem{ID: id, RequestPort: requestPort, LoggerPort: loggerPort, Alive: true}
}

func (ps *ProcessingSystem) Start() {
	go func() {
		for {
			req, ok := <-ps.RequestPort
			if !ok {
				break
			}
			startTimeProcessing := time.Now()
			req.WaitTime = startTimeProcessing.Sub(req.CreationTime())

			// delay := time.Microsecond * time.Duration(rand.Int63n(450)+50)
			delay := time.Millisecond
			startTime := time.Now()
			for time.Since(startTime) < delay {
			}

			endTimeProcessing := time.Now()
			req.TotalTime = endTimeProcessing.Sub(req.CreationTime())
			ps.LoggerPort <- logger.LogInfo{AuthorID: ps.ID, Req: req, CurrentTime: endTimeProcessing, OperationType: logger.Processed}
		}
		ps.Alive = false
	}()
}
