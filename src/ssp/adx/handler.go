package adx

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"ssp/dsp"
	_ "ssp/dsp/gdt"
	"ssp/protocol/adx"
	"ssp/util"
	"strconv"
	"strings"
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

	// Region filter
	if nil != slotConfig.Location {
		realIp := util.GetRealIp(req)
		regionFilter := true
		for _, r := range slotConfig.Location {
			if true == util.CheckIp4Region(realIp, r) {
				regionFilter = false
				break
			}
		}
		if true == regionFilter {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Request count limit
	adslotInfo := GetAdslotInfo(adslotId)
	if ok = checkReq(adslotId, adslotInfo, slotConfig); ok != true {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Smooth control
	if ok = smoothControl(adslotId, adslotInfo, slotConfig); ok != true {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	IncField(adslotId, preReq, slotConfig.EndDate)

	// Dsp handle
	handler, ok := dsp.HandlerMap[slotConfig.Dsp]
	if false == ok || nil == handler {
		util.Log.Error("The adslot has no dsp:%s, adslot id:%s", slotConfig.Dsp, adslotId)
		util.Log.Debug("HandlerMap info:%+v", dsp.HandlerMap)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res, err := handler.Handle(req, adxReq)
	if nil != err {
		util.Log.Error("%s, adslot id:%s", err.Error(), adslotId)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Response content filter
	if checkResContent(adslotId, res, slotConfig) != true {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if res.Ret == 0 {
		IncField(adslotId, preImp, slotConfig.EndDate)
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

func checkResContent(adslotId string, res *adx.Response, slotConfig *util.SlotConfig) bool {
	if nil == slotConfig.Filter {
		return true
	}
	for _, ad := range res.Data[adslotId] {
		for _, title := range slotConfig.Filter.Title {
			if strings.Contains(ad.Title, title) {
				util.Log.Debug("Title filter, adslot id:%s, response title:%s, filter title:%s",
					adslotId, ad.Title, title)
				return false
			}
		}
		for _, desc := range ad.Description {
			for _, fdesc := range slotConfig.Filter.Desc {
				if strings.Contains(desc, fdesc) {
					util.Log.Debug("Desc filter, adslot id:%s, response desc:%s, filter desc:%s",
						adslotId, desc, fdesc)
					return false
				}
			}
		}
		for _, imgUrl := range ad.ImgUrl {
			for _, fimgUrl := range slotConfig.Filter.Imageurl {
				if strings.Contains(imgUrl, fimgUrl) {
					util.Log.Debug("Image url filter, adslot id:%s, response url:%s, filter url:%s",
						adslotId, imgUrl, fimgUrl)
					return false
				}
			}
		}
	}
	return true
}
