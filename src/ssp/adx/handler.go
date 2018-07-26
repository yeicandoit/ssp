package adx

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"ssp/dsp"
	"ssp/protocol/adx"
	"ssp/util"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		req.Body.Close()
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			b := make([]byte, 1<<16)
			n := runtime.Stack(b, false)
			util.Log.Error("%s", b[:n])
		}
	}()

	if req.Method != http.MethodPost {
		util.Log.Error("Request method is:%s", req.Method) //设置log status,不打印此处return的日志
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		util.Log.Error("Read req.Body:%s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	adxReq := &adx.Request{}
	err = json.Unmarshal(body, adxReq)
	if err != nil {
		util.Log.Error("Unmarshal request body:%s, error:%s", string(body), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if nil == adxReq.Pos {
		util.Log.Error("adxReq.Pos is nil, request %+v", adxReq)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	adslotId := adxReq.Pos.Id
	slotConfig, ok := util.Adslot[adslotId]
	if false == ok || nil == slotConfig {
		util.Log.Error("The adslot does not exist, adslot id:%s", adslotId)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err = checkReq(adslotId, slotConfig); err != nil {
		util.Log.Error(err.Error())
		w.WriteHeader(http.StatusNoContent)
		return
	}
	handler, ok := dsp.HandlerMap[slotConfig.Dsp]
	if false == ok || nil == handler {
		util.Log.Error("The adslot has no dsp, adslot id:%s", adslotId)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res, err := handler(req, adxReq)
	if nil != err {
		util.Log.Error("%s, adslot id:%s", err.Error(), adslotId)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(res)
	if nil != err {
		util.Log.Error("%s, adslot id:%s", err.Error(), adslotId)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set(util.KHttpContentType, util.KHttpContentTypeJson)
	w.Header().Set(util.KHttpContentLength, strconv.Itoa(len(resp)))
	w.Write(resp)

}

func checkReq(adslotId string, slotConfig *util.SlotConfig) error {
	if nil == slotConfig {
		return errors.New("The slotConfig is nil, adslot id:%s", adslotId)
	}

	adslotInfo = GetAdslotInfo(adslotId)
	if nil == adslotInfo {
		return nil
	}
	if adslotInfo[makeFieldTotal(preReq)] > slotConfig.RequestTotal {
		return errors.New("Request total is over limit, adslot id:%s", adslotId)
	}
	if adslotInfo[makeField(preReq, 0)] > slotConfig.RequestDaily {
		return errors.New("Request daily is over limit, adslot id:%s", adslotId)
	}
	if adslotInfo[makeFieldTotal(preImp)] > slotConfig.ImpressionTotal {
		return errors.New("Impression total is over limit, adslot id:%s", adslotId)
	}
	if adslotInfo[makeField(preImp, 0)] > slotConfig.ImpressionDaily {
		return errors.New("Impression daily is over limit, adslot id:%s", adslotId)
	}
	IncField(adslotId, preReq, slotConfig.EndDate)
	return nil
}
