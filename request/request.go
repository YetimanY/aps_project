package request

import "time"

type Request struct {
	sourceID       uint64
	requestID      uint64
	creationTime   time.Time
	assetID        uint64
	assetPrice     uint64
	volumeOfAssets int64
	WaitTime       time.Duration
	TotalTime      time.Duration
}

func CreateRequest(sourceID, requestID, assetID, assetPrice uint64, volumeOfAssets int64) Request {
	if sourceID == 0 || requestID == 0 {
		return Request{}
	}
	tempRequest := Request{
		sourceID,
		requestID,
		time.Now(),
		assetID,
		assetPrice,
		volumeOfAssets,
		time.Duration(0),
		time.Duration(0),
	}
	return tempRequest
}

func (req Request) SourceID() uint64 {
	return req.sourceID
}

func (req Request) RequestID() uint64 {
	return req.requestID
}

func (req Request) CreationTime() time.Time {
	return req.creationTime
}

func (req Request) AssetID() uint64 {
	return req.assetID
}

func (req Request) AssetPrice() uint64 {
	return req.assetPrice
}

func (req Request) VolumeOfAssets() int64 {
	return req.volumeOfAssets
}
