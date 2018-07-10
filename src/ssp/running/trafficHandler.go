package running

import (
	"adx/util"
	"net/http"
	"strings"
)

// curl http://127.0.0.1:9090/traffic/req?sid=ifly&slotid=12345
// sh14 sh15 sh17 sh18 bj46 bj47
func QueryTrafficHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Real-Ip") != "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	queryPath := r.URL.Path
	queryPathItems := strings.Split(strings.Trim(queryPath, util.KSepSlash), util.KSepSlash)
	if len(queryPathItems) < 2 {
		http.Error(w, "path incomplete", http.StatusNoContent)
		return
	}

	_req := r.URL.Query()

	trafficLog.Sid = _req.Get("sid")
	trafficLog.SlotId = _req.Get("slotid")
	trafficLog.Size = _req.Get("size")
	trafficLog.SlotType = _req.Get("slottype")
	trafficLog.Cid = _req.Get("cid")
	trafficLog.LogType = queryPathItems[1]

	trafficLog.Reset()
	trafficLog.Run(w)
}
