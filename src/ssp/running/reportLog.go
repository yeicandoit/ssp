package running

import (
	"adx/dsp"
	"adx/proto/baidu"
	"adx/proto/ftx"
	"adx/proto/iax"
	"adx/proto/qax"
	"adx/proto/vam"
	"adx/proto/zplay"
	"adx/util"
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net/http"
	"strconv"
	"time"
)

const (
	typeReq  = "req"
	typeResp = "rsp"

	maxTimeWait int = 15
	maxRows     int = 15
)

type DebugLog struct {
	isRunning    bool
	finished     bool
	curRows      int
	resultBuffer *bytes.Buffer
}

type TrafficLog struct {
	*DebugLog
	Sid      string
	SlotId   string
	Size     string
	SlotType string
	Cid      string
	LogType  string
}

var trafficLog *TrafficLog

func init() {
	trafficLog = &TrafficLog{
		DebugLog: NewDebugLog(),
	}
}

func NewDebugLog() *DebugLog {
	return &DebugLog{
		isRunning:    false,
		finished:     false,
		curRows:      0,
		resultBuffer: bytes.NewBufferString(""),
	}
}

func SaveReqLog(req *dsp.BidRequest, contentType, sid string, body []byte) {
	if !trafficLog.isRunning {
		return
	}

	if trafficLog.LogType != typeReq {
		return
	}

	if trafficLog.doReqFilter(req) {
		return
	}

	switch contentType {
	case util.KHttpContentTypeJson:
		trafficLog.AppendResult(string(body))
	case util.KHttpContentTypeProto:
		pbStruct := getPbMessage(typeReq, sid)
		if pbStruct == nil {
			trafficLog.AppendResult(string(body))
		} else {
			if err := proto.Unmarshal(body, pbStruct); err != nil {
				trafficLog.AppendResult(err.Error())
			} else {
				trafficLog.AppendResult(fmt.Sprintf("%+v", pbStruct))
			}
		}
	}
}

func SaveRspLog(req *dsp.BidRequest, rspDsps *dsp.Winner, contentType, sid string, body []byte) {
	if !trafficLog.isRunning {
		return
	}

	if trafficLog.LogType != typeResp {
		return
	}

	if trafficLog.doRspFilter(rspDsps, req) {
		return
	}

	switch contentType {
	case util.KHttpContentTypeJson:
		trafficLog.AppendResult(string(body))
	case util.KHttpContentTypeProto:
		pbStruct := getPbMessage(typeResp, sid)
		if pbStruct == nil {
			trafficLog.AppendResult(string(body))
		} else {
			if err := proto.Unmarshal(body, pbStruct); err != nil {
				trafficLog.AppendResult(err.Error())
			} else {
				trafficLog.AppendResult(fmt.Sprintf("%+v", pbStruct))
			}
		}
	}
}

func (d *DebugLog) AppendResult(s string) {
	if d.finished {
		return
	}

	if d.curRows > maxRows {
		// 行数满了返回
		d.finished = true
		return
	}
	d.curRows++

	d.resultBuffer.WriteString(s + "\n")
}

func (d *DebugLog) Run(w http.ResponseWriter) {
	d.isRunning = true
	var timeCount int
	for {
		if timeCount > maxTimeWait || d.finished {
			break
		}
		timeCount++
		time.Sleep(time.Second)
	}
	// 时间到了返回
	d.finished = true
	if d.curRows == 0 {
		w.Write([]byte("empty result\n"))
	} else {
		w.Write(d.resultBuffer.Bytes())
	}
	d.resultBuffer.Reset()
	d.isRunning = false
}

func (d *DebugLog) Reset() {
	d.finished = false
	d.curRows = 0
}

func (t *TrafficLog) doReqFilter(req *dsp.BidRequest) bool {
	if req.GetSid() != t.Sid && t.Sid != "" {
		return true
	}

	if t.SlotId != "" {
		slotIdMatched := false
		for _, slot := range req.Imps {
			if slot.GetSlotId() == t.SlotId {
				slotIdMatched = true
				break
			}
		}
		if !slotIdMatched {
			return true
		}
	}

	if t.Size != "" {
		sizeMatched := false
		for _, slot := range req.Imps {
			if slot.Banner != nil {
				if t.Size == strconv.Itoa(int(slot.Banner.GetSlotSize().GetW()))+"x"+strconv.Itoa(int(slot.Banner.GetSlotSize().GetH())) {
					sizeMatched = true
					break
				}
			}
			if slot.Video != nil {
				if t.Size == strconv.Itoa(int(slot.Video.GetSlotSize().GetW()))+"x"+strconv.Itoa(int(slot.Video.GetSlotSize().GetH())) {
					sizeMatched = true
					break
				}
			}
			if slot.Native != nil {
				// 查主图尺寸
				if t.Size == strconv.Itoa(int(slot.Native.Asset[0].Image.GetW()))+"x"+strconv.Itoa(int(slot.Native.Asset[0].Image.GetH())) {
					sizeMatched = true
					break
				}
			}
		}
		if !sizeMatched {
			return true
		}
	}

	if t.SlotType != "" {
		slotTypeMatched := false
		for _, slot := range req.Imps {
			if t.SlotType == strconv.Itoa(int(slot.GetSlotType())) {
				slotTypeMatched = true
				break
			}
		}
		if !slotTypeMatched {
			return true
		}
	}

	return false
}

func (t *TrafficLog) doRspFilter(rspDsps *dsp.Winner, req *dsp.BidRequest) bool {
	if t.Cid != "" {
		for _, imp := range req.Imps {
			iid := imp.GetId()
			for _, v := range rspDsps.Imps[iid] {
				if t.Cid == strconv.FormatInt(v.Bid.GetAdid(), 10) {
					return false
				}
			}
		}
	}
	return true
}

func getPbMessage(logType, sid string) proto.Message {
	switch sid {
	case util.Rtype_Ftx:
		if logType == typeReq {
			return &ftx.BidRequest{}
		}
		if logType == typeResp {
			return &ftx.BidResponse{}
		}
	case util.Rtype_Baidu:
		if logType == typeReq {
			return &baidu.BidRequest{}
		}
		if logType == typeResp {
			return &baidu.BidResponse{}
		}
	case util.Rtype_Iax:
		if logType == typeReq {
			return &iax.BidRequest{}
		}
		if logType == typeResp {
			return &iax.BidResponse{}
		}
	case util.Rtype_QAX:
		if logType == typeReq {
			return &qax.BidRequest{}
		}
		if logType == typeResp {
			return &qax.BidResponse{}
		}
	case util.RType_VAM:
		if logType == typeReq {
			return &vam.VamRequest{}
		}
		if logType == typeResp {
			return &vam.VamResponse{}
		}
	case util.Rtype_Zplay:
		if logType == typeReq {
			return &zplay.BidRequest{}
		}
		if logType == typeResp {
			return &zplay.BidResponse{}
		}
	}
	return nil
}
