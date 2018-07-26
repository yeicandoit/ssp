package main

import (
	"net/http"
	"runtime"
	"ssp/adx"
	"ssp/util"
	"strconv"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := strconv.Itoa(util.ZsspServerPort)
	http.HandleFunc("/adx", adx.Handler)
	util.Log.Info("zsspserver starts listen :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	util.Log.Info("%v\n", err)
}
