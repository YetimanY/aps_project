package logger

import (
	"aps/request"
	"time"
)

const (
	Created = iota
	Added
	NotAdded
	Received
	Evicted
	Processing
	Processed
	SystemRunning
	SystemStopped
)

const (
	MesSystemRunning         = "[%v]: System is running.\n"
	MesSystemStopped         = "[%v]: System was stopped after %v of operation.\n"
	MesRequestCreated        = "[%v]: Request(%v:%v) was created.\n                   "
	MesRequestAdded          = "[%v]: Request(%v:%v) was added to buffer (index=%v).\n                   "
	MesRequestNotAdded       = "[%v]: Request(%v:%v) was not added to buffer (buffer is full).\n                   "
	MesRequestReceived       = "[%v]: Request(%v:%v) was received from buffer.\n                   "
	MesRequestEvicted        = "[%v]: Request(%v:%v) has been evicted from buffer.\n                   "
	MesRequestSentProcessing = "[%v]: Request(%v:%v) has been sent for processing (PS%v: busy).\n                   "
	MesRequestProcessed      = "[%v]: Request(%v:%v) was processed after %v %v in system (PS%v: free).\n                   "
	TimeFormat               = "15:04:05.000000"
)

type LogInfo struct {
	AuthorID      uint64
	Req           request.Request
	CurrentTime   time.Time
	OperationType int8
}

type Logger chan LogInfo
