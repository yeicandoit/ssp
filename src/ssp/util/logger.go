package util

import (
	//"bytes"
	"fmt"
)

type Bid_Status int

const (
	Bid_Status_Init      Bid_Status = 0
	Bid_Status_Bidding              = 1
	Bid_Status_Nobid                = 2
	Bid_Status_SspRead              = -1
	Bid_Status_SspParse             = -2
	Bid_Status_SspVerify            = -3
	Bid_Status_SspBuild             = -4
	Bid_Status_DspReq               = -5
	Bid_Status_SspResp              = -6
)

func idTrunc(id string) string {
	if len(id) > 6 {
		return id[len(id)-6:]
	}
	return id
}

func NewBidLogger(sid string) *BidLogger {
	return &BidLogger{
		sid:  sid,
		bids: make(map[int32]int8),
	}
}

type impLogger struct {
	slotid string
	dsps   map[int32][]string
}

type BidLogger struct {
	sid    string
	bid    string
	ext    string
	bids   map[int32]int8
	slots  map[string]impLogger
	errs   map[Bid_Status]error
	status Bid_Status
}

func (log *BidLogger) SetBid(bid string) {
	log.bid = idTrunc(bid)
}

func (log *BidLogger) AddDspRequest(impid string, slotid string, dspid int32) {
	impid = idTrunc(impid)
	slotid = idTrunc(slotid)
	if log.slots == nil {
		log.slots = make(map[string]impLogger)
	}
	if imp, ok := log.slots[impid]; ok {
		imp.dsps[dspid] = []string{}
	} else {
		log.slots[impid] = impLogger{
			slotid: slotid,
			dsps:   map[int32][]string{dspid: []string{}},
		}
	}
}

func (log *BidLogger) AddDspResponse(impid string, dspid int32, adid int64, price int32) {
	if log.slots == nil {
		return
	}

	impid = idTrunc(impid)
	if imp, ok := log.slots[impid]; ok {
		imp.dsps[dspid] = append(imp.dsps[dspid], fmt.Sprintf("%d-%d", adid, price))
	}

	log.bids[dspid] = 1
}

func (log *BidLogger) AddDspNobids(dspid int32) {
	log.bids[dspid] = 0
}

func (log *BidLogger) SetStatus(status Bid_Status, err error, ext string) {
	if err != nil {
		if log.errs == nil {
			log.errs = make(map[Bid_Status]error)
		}
		log.errs[status] = err
	}
	log.status = status
	log.ext = ext
}

func (log *BidLogger) Log() {

	if log.status == Bid_Status_Bidding {
		//Log.Info("[%s][%s][BID] %v", log.sid, log.bid, log.slots)
	} else if log.status == Bid_Status_Nobid {
		//Log.Info("[%s][%s][NBD] %v %v", log.sid, log.bid, log.bids, log.slots)
	} else {
		Log.Error("[%s][%s][S=%d][Ext=%s] %+v %+v", log.sid, log.bid, log.status, log.ext, log.errs, log.slots)
	}
}
