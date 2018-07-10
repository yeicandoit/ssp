package main

import (
	"adx/util"
	"net/http"
	"runtime"
	"strconv"
)

var handlerMap = map[string]http.HandlerFunc{
	"/im":       handler.ImHandler,
	"/ck":       handler.CkHandler,
	"/innersdk": innersdkhandler.InsdkHandler,
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := strconv.Itoa(util.ZAdxServerPort)
	util.InitIndex()
	for p, h := range handlerMap {
		http.HandleFunc(p, h)
	}

	http.Handle("/wax", wax.New())

	util.Log.Info("zadxserver starts listen :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	util.Log.Info("%v\n", err)
}
