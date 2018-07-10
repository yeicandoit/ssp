package gdt

import (
    "ssp/dsp/gdt"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var dspclient = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   500 * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost: 10,
	},
	Timeout: 500 * time.Millisecond,
}

var pos = &gdt.Pos{
	Id: 6000425899956957, //广告位ID
	// Width:  640,
	// Height: 960,
	SupportFullScreenInterstitial: false,
	AdCount:        1,
	NeedRenderedAd: true,
}

var media = &gdt.Media{
	AppId:       "1104238667",   //app_id
	AppBunbleId: "com.test.ios", //app_bundle_id
}

var device = &gdt.Device{
	Os:           "ios",
	OsVersion:    "9.3.2",
	Model:        "iPhone4,1",
	Manufacturer: "Apple",
	DeviceType:   1,
	Idfa:         "E2DFA890-496A-47FD-9941-DF1FC4E6484A",
	// Imei:         "011472001975695",
	// ImeiMd5:      "63ba33737bd172dad3919d49fc9f920f",
	// AndroidId:    "9774d56d682e549c",
	// AndroidIdMd5: "cf95dc53f383f9a836fd749f3ef439cd",
}

var network = &gdt.Network{
	ConnectType: 1,
	Carrier:     1,
}

func TestGdtHandler(t *testing.T) {
	jpos, _ := json.Marshal(pos)
	spos := string(jpos[:])
	jmedia, _ := json.Marshal(media)
	smedia := string(jmedia[:])
	jdevice, _ := json.Marshal(device)
	sdevice := string(jdevice[:])
	jnetwork, _ := json.Marshal(network)
	snetwork := string(jnetwork[:])

	req := "http://127.0.0.1:9092/gdt?api_version=3.0&support_https=0&pos=" + url.QueryEscape(spos) + "&media=" + url.QueryEscape(smedia) + "&device=" + url.QueryEscape(sdevice) + "&network=" + url.QueryEscape(snetwork)
	treq, _ := http.NewRequest("GET", req, nil)
	treq.Header.Set("X-Forwarded-For", "116.226.34.237")
	treq.Header.Set("User-Agent", "Golang")
	treq.Header.Set("Referer", "Nothing")
	gres, _ := dspclient.Do(treq)
	body, _ := ioutil.ReadAll(gres.Body)
	t.Log(string(body[:]))
}
