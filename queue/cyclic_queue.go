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
	buffer           []request.Request
	head, tail, size int
	mutex            sync.Mutex
	logger           logger.Logger
}

func (queue *CyclicQueue) Tail() int {
	return queue.tail
}

func (queue *CyclicQueue) Capacity() int {
	return cap(queue.buffer)
}

func (queue *CyclicQueue) IsFull() bool {
	return queue.size == cap(queue.buffer)
}

func (queue *CyclicQueue) IsEmpty() bool {
	return queue.size == 0
}

func CreateCyclicQueue(capacity int, log logger.Logger) *CyclicQueue {
	return &CyclicQueue{
		buffer: make([]request.Request, capacity),
		head:   0,
		tail:   0,
		size:   0,
		mutex:  sync.Mutex{},
		logger: log,
	}
}

func (q *CyclicQueue) AddRequest(req request.Request) request.Request {
	for !q.mutex.TryLock() {
	}
	defer q.mutex.Unlock()

	returnReq := req
	if q.IsFull() {
		if q.logger != nil {
			q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.NotAdded}
		}
		returnReq = q.buffer[q.head]
		q.head = (q.head + 1) % cap(q.buffer)
		if q.logger != nil {
			q.logger <- logger.LogInfo{Req: returnReq, CurrentTime: time.Now(), OperationType: logger.Evicted}
		}
	} else {
		q.size++
	}
	q.buffer[q.tail] = req
	q.tail = (q.tail + 1) % cap(q.buffer)
	if q.logger != nil {
		q.logger <- logger.LogInfo{Req: req, CurrentTime: time.Now(), OperationType: logger.Added}
	}
	return returnReq
}

func (q *CyclicQueue) GetRequest() request.Request {
	for !q.mutex.TryLock() {
	}
	defer q.mutex.Unlock()

	if q.IsEmpty() {
		return request.Request{}
	}
	returnReq := q.buffer[q.head]
	q.head = (q.head + 1) % cap(q.buffer)
	q.size--
	if q.logger != nil {
		q.logger <- logger.LogInfo{Req: returnReq, CurrentTime: time.Now(), OperationType: logger.Received}
	}
	return returnReq
}

func (q *CyclicQueue) StringQueueStatus() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Buffer[%v/%v]: [", q.size, cap(q.buffer)))
	for i := 0; i < len(q.buffer); i++ {
		if q.buffer[i].SourceID() != 0 {
			sb.WriteString(fmt.Sprintf(" %v:%v", q.buffer[i].SourceID(), q.buffer[i].RequestID()))
		} else {
			sb.WriteString(" -")
		}
	}
	sb.WriteString(" ].\n                   ")
	return sb.String()
}
