package chushou

import (
	"adx/util"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const Gdt_url = "http://mi.gdt.qq.com/api/v3"

var dspclient = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   1000 * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost: 10,
	},
	Timeout: 1000 * time.Millisecond,
}

func ChushouHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			b := make([]byte, 1<<16)
			n := runtime.Stack(b, false)
			util.Log.Error("[chushou-panic]%s", b[:n])
		}
	}()
	req_ := Gdt_url + "?" + r.URL.RawQuery
	treq, err := http.NewRequest("GET", req_, nil)
	if err != nil {
		util.Log.Error("[chushou] http.NewRequest:%s", err.Error())
		responseNobid(w)
		return
	}
	treq.Header.Set("X-Forwarded-For", util.GetRealIp(r))
	treq.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	treq.Header.Set("Referer", r.Header.Get("Referer"))
	gres, err := dspclient.Do(treq)
	if err != nil || gres == nil {
		responseNobid(w)
		return
	}
	defer gres.Body.Close()
	body, _ := ioutil.ReadAll(gres.Body)
	w.Header().Set(util.KHttpContentType, util.KHttpContentTypeJson)
	w.Header().Set(util.KHttpContentLength, strconv.Itoa(len(body)))
	w.WriteHeader(gres.StatusCode)
	w.Write(body)
}

func responseNobid(w http.ResponseWriter) {
	resp := &GdtResponse{}
	bt, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	w.Header().Set(util.KHttpContentType, util.KHttpContentTypeJson)
	w.Header().Set(util.KHttpContentLength, strconv.Itoa(len(bt)))
	w.Write(bt)
}
