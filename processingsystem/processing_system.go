package processingsystem

import (
	"aps/logger"
	"aps/request"
	"fmt"
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
	var a time.Duration
	go func() {
		for {
			st := time.Now()
			for ps.status == 0 && !ps.interrupt {
			}
			if ps.interrupt {
				break
			}
			req := ps.buffer
			b := time.Since(st)
			fmt.Println(b)
			a += b
			startTimeProcessing := time.Now()
			req.WaitTime = startTimeProcessing.Sub(req.CreationTime())

			// delay := time.Microsecond * time.Duration(rand.Int63n(450)+50)
			delay := time.Millisecond * 2
			startTime := time.Now()
			for time.Since(startTime) < delay {
			}

			endTimeProcessing := time.Now()
			ps.status = 0
			req.TotalTime = endTimeProcessing.Sub(req.CreationTime())
			ps.loggerPort <- logger.LogInfo{AuthorID: ps.id, Req: req, CurrentTime: endTimeProcessing, OperationType: logger.Processed}
		}
		fmt.Println(ps.id, a)
		ps.status = -1
	}()
}

func (ps *ProcessingSystem) SendRequest(req request.Request) {
	ps.buffer = req
	// fmt.Println(req)
	ps.status = 1
}

func (ps *ProcessingSystem) Interrupt() {
	ps.interrupt = true
}

func (ps *ProcessingSystem) Status() int {
	return ps.status
}
