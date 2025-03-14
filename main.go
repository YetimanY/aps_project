package main

import (
	"aps/logger"
	"aps/processingsystem"
	"aps/productionmanager"
	"aps/queue"
	"aps/request"
	"aps/selectionmanager"
	"aps/source"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	numberSources = 3
	numberDevices = 5
	bufferSize    = 25
	lambda        = 3.5
	numberRequest = 2500
)

func main() {
	closedFlag1 := atomic.Value{}
	closedFlag1.Store(false)
	closedFlag2 := atomic.Value{}
	closedFlag2.Store(false)
	log := make(logger.Logger, numberRequest*2)
	wg := sync.WaitGroup{}
	buf := queue.CreateCyclicQueue(bufferSize, log)
	port := make(chan request.Request)

	StartLogger(log, bufferSize, numberSources, numberDevices, &wg)

	pss := make([]processingsystem.ProcessingSystem, numberDevices)
	for i := 0; i < numberDevices; i++ {
		pss[i] = *processingsystem.CreateProcessingSystem(uint64(i+1), log)
	}
	sm := selectionmanager.CreateSM(pss, buf, log, &closedFlag2, &wg)
	pm := productionmanager.CreatePM(port, buf, log, &wg)
	sourcePool := make([]source.Source, numberSources)
	for i := 0; i < numberSources; i++ {
		sourcePool[i] = *source.CreateSource(uint64(i+1), lambda, port, log, &wg, &closedFlag1)
	}

	log <- logger.LogInfo{CurrentTime: time.Now(), OperationType: logger.SystemRunning}
	for i := 0; i < numberDevices; i++ {
		pss[i].Start()
	}
	sm.Start()
	pm.Start()
	for i := 0; i < numberSources; i++ {
		sourcePool[i].Start()
	}
	for pm.NumRequest() <= uint64(numberRequest) {

	}
	wg.Add(numberSources)
	closedFlag1.Store(true)
	wg.Wait()

	wg.Add(1)
	close(port)
	wg.Wait()

	wg.Add(1)
	closedFlag2.Store(true)
	wg.Wait()

	wg.Add(1)
	log <- logger.LogInfo{CurrentTime: time.Now(), OperationType: logger.SystemStopped}
	wg.Wait()
}

func StartLogger(log logger.Logger, bufferSize int, numSources int, numDevices int, wg *sync.WaitGroup) {
	go func() {
		buffer := queue.CreateCyclicQueue(bufferSize, nil)
		bufferState := buffer.StringState()

		evictReqs := make([]int, numSources)
		processedReq := make([][]request.Request, numSources)

		deviceMes := "{PS%v(%v:%v): %v }"
		devices := make([]string, numDevices)
		for i := 0; i < numDevices; i++ {
			devices[i] = fmt.Sprintf(deviceMes, i+1, 0, 0, "free")
		}
		devicesTime := make([]time.Duration, numDevices)
		devicesState := StringDevicesState(devices)

		offsetString := "\n                   "
		var (
			systemStartTime time.Time
			systemEndTime   time.Time
		)
		sb := strings.Builder{}

	loop:
		for {
			info := <-log

			authorID := info.AuthorID
			req := info.Req
			curTime := info.CurrentTime
			operationType := info.OperationType

			switch operationType {
			case logger.Created:
				sb.WriteString(fmt.Sprintf(logger.MesRequestCreated, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID()))
			case logger.Added:
				sb.WriteString(fmt.Sprintf(logger.MesRequestAdded, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID(), ((buffer.Tail())+buffer.Capacity())%buffer.Capacity()))
				buffer.AddRequest(req)
				bufferState = buffer.StringState()
			case logger.NotAdded:
				sb.WriteString(fmt.Sprintf(logger.MesRequestNotAdded, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID()))
			case logger.Received:
				sb.WriteString(fmt.Sprintf(logger.MesRequestReceived, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID()))
				buffer.GetRequest()
				bufferState = buffer.StringState()
			case logger.Evicted:
				sb.WriteString(fmt.Sprintf(logger.MesRequestEvicted, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID()))
				buffer.GetRequest()
				bufferState = buffer.StringState()
				evictReqs[req.SourceID()-1] += 1
			case logger.Processing:
				sb.WriteString(fmt.Sprintf(logger.MesRequestSentProcessing, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID(), authorID))
				devices[authorID-1] = fmt.Sprintf(deviceMes, authorID, req.SourceID(), req.RequestID(), "busy")
				devicesState = StringDevicesState(devices)
			case logger.Processed:
				sb.WriteString(fmt.Sprintf(logger.MesRequestProcessed, curTime.Format(logger.TimeFormat), req.SourceID(), req.RequestID(), req.TotalTime, authorID))
				devices[authorID-1] = fmt.Sprintf(deviceMes, authorID, 0, 0, "free")
				devicesState = StringDevicesState(devices)
				devicesTime[authorID-1] += req.TotalTime - req.WaitTime
				processedReq[req.SourceID()-1] = append(processedReq[req.SourceID()-1], req)
			case logger.SystemRunning:
				systemStartTime = curTime
				fmt.Println(fmt.Sprintf(logger.MesSystemRunning, curTime.Format(logger.TimeFormat)))
				fmt.Println("----------------------------------------------------------------------------------------------------------")
				continue loop
			case logger.SystemStopped:
				systemEndTime = curTime
				totalTimeSystem := systemEndTime.Sub(systemStartTime)
				fmt.Println(fmt.Sprintf(logger.MesSystemStopped, curTime.Format(logger.TimeFormat), totalTimeSystem))
				fmt.Println()
				fmt.Println(FinalStatistics(devicesTime, totalTimeSystem, processedReq, evictReqs))
				break loop
			}
			sb.WriteString(offsetString)
			sb.WriteString(bufferState)
			sb.WriteString(offsetString)
			sb.WriteString(devicesState)
			sb.WriteString("----------------------------------------------------------------------------------------------------------")
			// fmt.Println(sb.String())
			sb.Reset()
		}
		wg.Done()
	}()
}

func StringDevicesState(devices []string) string {
	sb := strings.Builder{}
	sb.WriteString("Devices: [")
	for i := 0; i < len(devices); i++ {
		sb.WriteString(" ")
		sb.WriteString(devices[i])
	}
	sb.WriteString(" ].\n")
	return sb.String()
}

func FinalStatistics(devices []time.Duration, totalTime time.Duration, processedReq [][]request.Request, evictReq []int) string {
	sb := strings.Builder{}
	sb.WriteString("Sources statistics\n")
	sb.WriteString(fmt.Sprintf("%-14s|%-14s|%-14s|%-18s|%-14s|%-16s|%-18s|%-18s|%-18s|\n", "Source ID", "Total Req", "Evicted Req", "Eviction Factor", "Avg T(total)", "Avg T(buffer)", "Avg T(processed)", "Disp(buffer)", "Disp(processed)"))
	totalReq := 0
	totalEvict := 0

	for i := 0; i < len(processedReq); i++ {
		var (
			totalReqTime     uint64
			bufferReqTime    uint64
			processedReqTime uint64
			dispBuffer       uint64
			dispProcessed    uint64
		)
		for j := 0; j < len(processedReq[i]); j++ {
			totalReqTime += uint64(processedReq[i][j].TotalTime.Microseconds())
			bufferReqTime += uint64(processedReq[i][j].WaitTime.Microseconds())
			processedReqTime += uint64((processedReq[i][j].TotalTime - processedReq[i][j].WaitTime).Microseconds())
		}
		avgTotal := int(totalReqTime) / len(processedReq[i])
		avgBuffer := int(bufferReqTime) / len(processedReq[i])
		avgProcessed := int(processedReqTime) / len(processedReq[i])
		for j := 0; j < len(processedReq[i]); j++ {
			a := (processedReq[i][j].WaitTime.Milliseconds() - int64(avgBuffer/1000))
			dispBuffer += uint64(a * a)

			aa := ((processedReq[i][j].TotalTime - processedReq[i][j].WaitTime).Microseconds() - int64(avgProcessed))
			dispProcessed += uint64(aa * aa)
		}

		iTotalReq := len(processedReq[i]) + evictReq[i]
		totalReq += iTotalReq
		totalEvict += evictReq[i]
		fAvgTotal := fmt.Sprintf("%.2fms", float64(avgTotal)/1000.0)
		fAvgBuffer := fmt.Sprintf("%.2fms", float64(avgBuffer)/1000.0)
		fAvgProcessed := fmt.Sprintf("%.2fms", float64(avgProcessed)/1000.0)
		dispB := fmt.Sprintf("%dms", dispBuffer/uint64(len(processedReq[i])))
		dispP := fmt.Sprintf("%dµs", dispProcessed/uint64(len(processedReq[i])))
		evictionFactor := float64(evictReq[i]) / float64(iTotalReq)
		mes := "%-14v|%-14v|%-14v|%-18.2f|%-14v|%-16v|%-18v|%-18v|%-18v|\n"
		sb.WriteString(fmt.Sprintf(mes, i+1, iTotalReq, evictReq[i], evictionFactor, fAvgTotal, fAvgBuffer, fAvgProcessed, dispB, dispP))
	}
	sb.WriteString("---------------------------------------------------------------------------------------------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("%-14v|%-14v|%-14v|%-18.2f|\n", "Total", totalReq, totalEvict, float64(totalEvict)/float64(totalReq)))
	sb.WriteString("----------------------------------------------------------------\n\n")
	sb.WriteString("Devices statistics\n")
	sb.WriteString(fmt.Sprintf("%-14s|%-20s|%-20s|\n", "Device ID", "Utilization Factor", "Total Active Time"))
	for i := 0; i < cap(devices); i++ {
		sb.WriteString(fmt.Sprintf("%-14v|%-20.2f|%-20v|\n", i+1, float64(devices[i])/float64(totalTime), devices[i]))
	}
	return sb.String()
}
