package queue

import (
	"aps/logger"
	"aps/request"
	"fmt"
	"strings"
	"sync"
	"time"
)

type CyclicQueue struct {
	mutex    sync.Mutex
	requests []request.Request
	front    int
	size     int
	logger   logger.Logger
}

func (q *CyclicQueue) getRequest() request.Request {
	if q.size == 0 {
		return request.Request{}
	}
	req := q.requests[q.front]
	q.requests[q.front] = request.Request{}
	q.front = (q.front + 1) % cap(q.requests)
	q.size -= 1
	return req
}

func CreateCyclicQueue(size int, log logger.Logger) *CyclicQueue {
	tempBuff := make([]request.Request, size)
	queue := CyclicQueue{sync.Mutex{}, tempBuff, 0, 0, log}
	return &queue
}

func (q *CyclicQueue) AddRequest(req request.Request) bool {
	q.mutex.Lock()

	if q.size == cap(q.requests) {
		if q.logger != nil {
			q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.NotAdded}
		}
		return false
	} else {
		rear := (q.front + q.size) % cap(q.requests)
		q.requests[rear] = req
		q.size += 1
		if q.logger != nil {
			q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.Added}
		}
		q.mutex.Unlock()
		return true
	}
}

func (q *CyclicQueue) GetRequest() request.Request {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	req := q.getRequest()
	if req.SourceID() == 0 {
		return req
	}
	if q.logger != nil {
		q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.Received}
	}
	return req
}

func (q *CyclicQueue) EvictRequest() request.Request {
	defer q.mutex.Unlock()

	req := q.getRequest()
	if q.logger != nil {
		q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.Evicted}
	}
	return req
}

func (q *CyclicQueue) Rear() int {
	return (q.front + q.size) % cap(q.requests)
}

func (q *CyclicQueue) StringQueueStatus() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Buffer[%v/%v]: [", q.size, cap(q.requests)))
	for i := 0; i < len(q.requests); i++ {
		if q.requests[i].SourceID() != 0 {
			sb.WriteString(fmt.Sprintf(" %v:%v", q.requests[i].SourceID(), q.requests[i].RequestID()))
		} else {
			sb.WriteString(" -")
		}
	}
	sb.WriteString(" ].\n                   ")
	return sb.String()
}

func (q *CyclicQueue) Size() int {
	return q.size
}
