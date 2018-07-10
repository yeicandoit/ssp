//go:generate encode -type=ReportExt
package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"thrift-gen/noticeserver"
	"time"
	"tripod/define"
	"tripod/devkit"
	"tripod/zkutil"

	"git.apache.org/thrift.git/lib/go/thrift"
)

func GenRandId() string {
	s := fmt.Sprintf("%d%s", time.Now().Unix(), devkit.RandString(6))
	return devkit.HashFunc(s)
}

//build safe info
func BuildSafeUrl(info string) string {
	return url.QueryEscape(base64.URLEncoding.EncodeToString([]byte(info)))
}

type ReportExt struct {
	Sid         string `tag:"sid"`
	AppId       string `tag:"appid"`
	SlotType    int32  `tag:"slotType"`
	SlotId      string `tag:"slotid"`
	Bidfloor    int32  `tag:"bidfloor"`
	SettlePrice int32  `tag:"price"`
	DevEarn     int32  `tag:"devearn"`
	Muid        string `tag:"muid"`
	IdType      int32  `tag:"idType"`
	BidType     int32  `tag:"bidType"`
	Ct          int32  `tag:"ct"`
	Dt          int32  `tag:"dt"`
	Os          string `tag:"os"`
	Ip          string `tag:"ip"`
	Its         int64  `tag:"its"`
	Cts         int64  `tag:"cts"`
	Rts         int64  `tag:"rts"`
	Respts      int64  `tag:"respts"`
	Brand       string `tag:"brand"`
	Osv         string `tag:"osver"`
	Iid         string `tag:"iid"`
	DspId       int32  `tag:"dspid"`
	CheckSum    string `tag:"csum"`
	AdvId       int64  `tag:"advId"`
	AdId        int64  `tag:"adId"`
	RequestID   string `tag:"reqid"`
	RevenueMode int32  `tag:"revenueMode"`
	Status      uint32 `tag:"status"` //记录Dsp竞价响应状态
	Size        string `tag:"size"`   //流量尺寸
}

const (
	StatusAllOK         = 0  //竞价成功
	StatusBidFailed     = 1  //竞价失败
	StatusNoBid         = 2  //没有参与竞价
	StatusBidTimeout    = 4  //竞价超时
	StatusMateNoUseAudi = 8  //素材无效或未审核
	StatusNetErr        = 16 //网络有问题
)

func (p *ReportExt) Dump() string {
	info := ""
	v := reflect.ValueOf(*p)
	t := reflect.TypeOf(*p)
	fieldNo := t.NumField()
	for i := 0; i < fieldNo; i++ {
		fieldS := t.Field(i)
		fieldV := v.Field(i)
		if fieldS.Tag.Get("tag") == "" {
			continue
		}
		if i != 0 {
			info += "&"
		}

		info += fieldS.Tag.Get("tag") + "=" + fmt.Sprintf("%v", fieldV.Interface())
	}
	return info
}

func SendNotice(url string, header http.Header) {
	zkInstance := zkutil.GetZkInstance(ServiceConfig.ZkHosts)
	ts, err := zkInstance.GetOpenedTransport(define.GroupServing, define.NoticeServerName, true)
	if err != nil {
		Log.Error("[%d]get openedTransport for httpNotice err:%v", define.ErrCodeHttpNoticeNotConnect, err)
		return
	}
	defer ts.Close()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	noticeClient := noticeserver.NewNoticeClientFactory(ts, protocolFactory)
	_, err = noticeClient.PingWithHeader(url, header)
	if err != nil {
		Log.Warn("[%d]ping %s err:%v", define.ErrCodePingHttpNotice, url, err)
	}

}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

var realIpHeaders = [2]string{"X-FORWARDED-FOR", "X-REAL-IP"}

func GetRealIp(r *http.Request) (ip string) {
	for _, k := range realIpHeaders {
		if ip = r.Header.Get(k); ip != "" {
			ss := strings.Split(ip, ",")
			for _, s := range ss {
				if ip = strings.TrimSpace(s); ip != "" && ip != "unknown" {
					return
				}
			}
		}
	}

	if ip = r.RemoteAddr; ip != "" {
		ip = strings.Split(ip, ":")[0]
		return
	}

	return "0.0.0.0"
}

func GetAdvertiserDomain(aurl string, sid string) string {
	parseUrl, err := url.Parse(aurl)
	if err != nil {
		Log.Warn("%s parse doamin err:%s", sid, err.Error())
		return ""
	}
	dhost := strings.Split(parseUrl.Host, KSepDot)
	if l := len(dhost); l == 2 {
		return parseUrl.Host
	} else if l > 2 {
		if dhost[l-3] == "www" {
			return dhost[l-2] + KSepDot + dhost[l-1]
		} else {
			return dhost[l-3] + KSepDot + dhost[l-2] + KSepDot + dhost[l-1]
		}
	}
	return ""
}
