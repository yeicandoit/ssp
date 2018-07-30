package util

import (
	"flag"
	"path/filepath"
	"tripod/devkit"
	"tripod/zconf"

	l4g "tripod/3rdparty/code.google.com/p/log4go"
)

const ServiceConfigFile = "conf/ssp.yaml"
const AdslotConfigFile = "conf/adslot.yaml"

var Log l4g.Logger

var (
	rootPath       string
	ZsspServerPort int
	AdxRedisPool   *devkit.ZRedis
)

var ServiceConfig struct {
	ZsspServerLogConfigFile string
	GdtUrl                  string
	BaiduUrl                string
	AdxRedisAddress         []string
	Ipfile                  string
}

var Adslot = map[string]*SlotConfig{}

type SlotConfig struct {
	Dsp             string
	RequestTotal    int64
	RequestDaily    int64
	ImpressionTotal int64
	ImpressionDaily int64
	Location        []int64
	EndDate         string
	Filter          *Filter
}

type Filter struct {
	Title    []string
	Desc     []string
	Imageurl []string
}

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

func init() {
	parseArgs()
	err := zconf.ParseYaml(filepath.Join(rootPath, ServiceConfigFile), &ServiceConfig)
	if err != nil {
		panic(err)
	}
	err = zconf.ParseYaml(filepath.Join(rootPath, AdslotConfigFile), &Adslot)
	if err != nil {
		panic(err)
	}
	initIpCache()
	AdxRedisPool = devkit.NewZRedis(ServiceConfig.AdxRedisAddress)
	Log = devkit.NewLogger(devkit.GetAbsPath(ServiceConfig.ZsspServerLogConfigFile, rootPath))
	Log.Info("zsspserver config: %+v", ServiceConfig)
	Log.Info("adslot config: %+v", Adslot)
}
