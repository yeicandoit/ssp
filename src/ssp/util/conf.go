package util

import (
	"flag"
	"path/filepath"
	"storage"
	"time"
	"tripod/devkit"
	"tripod/zconf"

	l4g "tripod/3rdparty/code.google.com/p/log4go"
)

const ServiceConfigFile = "conf/adx.yaml"

var Log l4g.Logger

var (
	rootPath       string
	ZAdxServerPort int
	CheckInterval  time.Duration
)

func parseArgs() {
	flag.StringVar(&rootPath, "rootPath", "/opt/zyz/ssp", "Root Path")
	flag.IntVar(&ZAdxServerPort, "zadxserverport", 9090, "zsspserver listen port, range [9090, 9099]")
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
	ZadxServerLogConfigFile string
	ConfCheckInterval       int
	IndexDataFile           string
	IpDataPath              string
	ZkHosts                 string
	DefaultDspAddr          string
	Host                    string
	DspImMonitor            string
	DspCkMonitor            string
}

func init() {
	parseArgs()
	err := zconf.ParseYaml(filepath.Join(rootPath, ServiceConfigFile), &ServiceConfig)
	if err != nil {
		panic(err)
	}
	CheckInterval = time.Duration(ServiceConfig.ConfCheckInterval)
	Log = devkit.NewLogger(devkit.GetAbsPath(ServiceConfig.ZadxServerLogConfigFile, rootPath))
	Log.Info("zadxserver config: %+v", ServiceConfig)
}

func loadIndex(file string, args ...interface{}) (interface{}, bool) {
	newIndex := storage.NewIndex()
	newIndex.Load(file)

	Log.Info("idx len:publisher:%d;ssp_adspace:%d;ssp:%d;adspace:%d;material:%d;"+
		"material_audit:%d;app:%d;tactics:%d;dsp:%d;dsp_advertiser_audit:%d;autohome_info:%d",
		len(newIndex.Publishers), len(newIndex.SSPAdspaces), len(newIndex.Ssps),
		len(newIndex.AdSpaces), len(newIndex.Materials), len(newIndex.MaterialAudits),
		len(newIndex.Apps), len(newIndex.Tactics), len(newIndex.Dsps), len(newIndex.DspAdvertiserAudits),
		len(newIndex.AutohomeInfo))
	return newIndex, true
}

func InitIndex() {
	newIndex, _ := loadIndex(ServiceConfig.IndexDataFile)
	PAdxIndex = newIndex.(*storage.Index)
	go devkit.ReloadFile(loadIndex, ServiceConfig.IndexDataFile, LoadInterval, &PAdxIndex)
}
