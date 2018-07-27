package util

import (
	"bufio"
	"math/big"
	"net"
	"strconv"
	"strings"
)

type RegionIp struct {
	StartIp int64
	EndIp   int64
}

var Region2Ip map[int64][]*RegionIp

func initIpCache() {
	Region2Ip = make(map[int][]*RegionIp)
	fi, err := os.Open(rootPath + "/" + ServiceConfig.Ipfile)
	if err != nil {
		Log.Error("Open ip file:%s", err.Error())
		return
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		ipRegion := strings.Split(string(line), ",")
		if 3 != len(ipRegion) {
			Log.Error("ip error:%s", string(line))
			continue
		}
		regionCode, err := strconv.ParseInt(ipRegion[2], 10, 0)
		if err != nil {
			Log.Error("ip region code:%s", err.Error())
			continue
		}
		rip := &RegionIp{
			StartIp: InetAtoN(ipRegion[0]),
			EndIp:   InetAtoN(ipRegion[1]),
		}
		Region2Ip[regionCode] = append(Region2Ip[regionCode], rip)
	}
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func CheckIp4Region(ip string, regionCode int64) bool {
	ipcode := InetAtoN(ip)
	for _, r := range Region2Ip[regionCode] {
		if ipcode >= r.StartIp && ipcode <= r.EndIp {
			return true
		}
	}
	return false
}
