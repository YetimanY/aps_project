package processingsystem

import (
	"aps/logger"
	"aps/request"
	"time"
)

type ProcessingSystem struct {
	id         uint64
	buffer     request.Request
	loggerPort logger.Logger
	interrupt  bool
	status     int
}

func CreateProcessingSystem(id uint64, loggerPort logger.Logger) *ProcessingSystem {
	return &ProcessingSystem{id: id, buffer: request.Request{}, loggerPort: loggerPort, interrupt: false, status: 0}
}

func (ps *ProcessingSystem) Start() {
	go func() {
		for {
			for ps.status == 0 && !ps.interrupt {
			}
			if ps.interrupt {
				break
			}
			req := ps.buffer
			startTimeProcessing := time.Now()
			req.WaitTime = startTimeProcessing.Sub(req.CreationTime())

			delay := time.Millisecond * 10
			startTime := time.Now()
			for time.Since(startTime) < delay {
			}

			endTimeProcessing := time.Now()
			ps.status = 0
			req.TotalTime = endTimeProcessing.Sub(req.CreationTime())
			ps.loggerPort <- logger.LogInfo{AuthorID: ps.id, Req: req, CurrentTime: endTimeProcessing, OperationType: logger.Processed}
		}
		ps.status = -1
	}()
}

func (ps *ProcessingSystem) SendRequest(req request.Request) {
	ps.buffer = req
	ps.status = 1
}

func (ps *ProcessingSystem) Interrupt() {
	ps.interrupt = true
}

func (ps *ProcessingSystem) Status() int {
	return ps.status
}
