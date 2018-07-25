package util

import (
	"flag"
	"path/filepath"
	"time"
	"tripod/devkit"
	"tripod/zconf"

	l4g "tripod/3rdparty/code.google.com/p/log4go"
)

const ServiceConfigFile = "conf/ssp.yaml"

var Log l4g.Logger

var (
	rootPath       string
	ZsspServerPort int
	CheckInterval  time.Duration
)

func parseArgs() {
	flag.StringVar(&rootPath, "rootPath", "/opt/zyz/ssp", "Root Path")
	flag.IntVar(&ZsspServerPort, "zsspserverport", 9090, "zsspserver listen port, range [9090, 9099]")
	flag.Parse()

	if !filepath.IsAbs(rootPath) {
		p, err := filepath.Abs(rootPath)
		if err != nil {
			panic("Convert root path to abs path failed")
		}
		rootPath = p
	}
}

var ServiceConfig struct {
	ZsspServerLogConfigFile string
	ConfCheckInterval       int
	Host                    string
	DspImMonitor            string
	DspCkMonitor            string
	GdtUrl                  string
	BaiduUrl                string
}

func init() {
	parseArgs()
	err := zconf.ParseYaml(filepath.Join(rootPath, ServiceConfigFile), &ServiceConfig)
	if err != nil {
		panic(err)
	}
	CheckInterval = time.Duration(ServiceConfig.ConfCheckInterval)
	Log = devkit.NewLogger(devkit.GetAbsPath(ServiceConfig.ZsspServerLogConfigFile, rootPath))
	Log.Info("zsspserver config: %+v", ServiceConfig)
}
