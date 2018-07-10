package main

import (
    "ssp/dsp/gdt"
	"ssp/util"
	"net/http"
	"runtime"
	"strconv"
)

var handlerMap = map[string]http.HandlerFunc{
	"/gdt": gdt.GdtHandler,
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := strconv.Itoa(util.ZsspServerPort)
	for p, h := range handlerMap {
		http.HandleFunc(p, h)
	}

	util.Log.Info("zsspserver starts listen :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	util.Log.Info("%v\n", err)
}
