package source

import (
	"aps/logger"
	"aps/request"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type Source struct {
	id         uint64
	lambda     float64
	portPM     chan request.Request
	logger     logger.Logger
	wg         *sync.WaitGroup
	closedFlag *atomic.Value
}

func CreateSource(id uint64, lambda float64, portPM chan request.Request, logger logger.Logger, wg *sync.WaitGroup, closedFlag *atomic.Value) *Source {
	return &Source{id, lambda, portPM, logger, wg, closedFlag}
}

func (s *Source) Start() {
	go func() {
		for rID := 1; s.closedFlag.Load() == false; rID++ {
			delay := time.Microsecond * time.Duration(rand.ExpFloat64()/s.lambda*10000)
			startTime := time.Now()
			for time.Since(startTime) < delay {
			}
			req := request.CreateRequest(s.id, uint64(rID), 5, 5, 5)
			s.logger <- logger.LogInfo{Req: req, CurrentTime: req.CreationTime(), OperationType: logger.Created}
			s.portPM <- req
		}
		s.wg.Done()
	}()
}
