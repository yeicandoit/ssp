package gdt

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"ssp/dsp"
	"ssp/protocol/adx"
	"ssp/protocol/gdt"
	"ssp/util"
	"strconv"
)

const SidName = "gdt"

func init() {
	dsp.RegisterHandler(SidName, New())
}

type GdtHandler struct {
	dsp.BaseHandler
}

func New() *GdtHandler {
	h := &GdtHandler{}
	h.Delegate = h
	h.Sid = SidName
	return h
}

func (h *GdtHandler) SendDspRequest(r *http.Request, req *adx.Request) ([]byte, error) {
	pid, _ := strconv.ParseInt(req.Pos.Id, 10, 64)
	pos := &gdt.Pos{
		Id:     pid,
		Width:  req.Pos.Width,
		Height: req.Pos.Height,
		SupportFullScreenInterstitial: req.Pos.SupportFullScreenInterstitial,
		AdCount:        req.Pos.AdCount,
		NeedRenderedAd: req.Pos.NeedRenderedAd,
		LastAdIds:      req.Pos.LastAdIds,
		Channel:        req.Pos.Channel,
		PageNumber:     req.Pos.PageNumber,
	}
	jpos, _ := json.Marshal(pos)
	jmedia, _ := json.Marshal(req.Media)
	jdevice, _ := json.Marshal(req.Device)
	jnetwork, _ := json.Marshal(req.Network)
	sgeo := ""
	if nil != req.Geo {
		jgeo, _ := json.Marshal(req.Geo)
		sgeo = string(jgeo[:])
	}
	req_ := util.ServiceConfig.GdtUrl + "?api_version=" + req.ApiVersion
	req_ += "&support_https=" + strconv.Itoa(int(req.SupportHttps))
	req_ += "&pos=" + url.QueryEscape(string(jpos))
	req_ += "&media=" + url.QueryEscape(string(jmedia))
	req_ += "&device=" + url.QueryEscape(string(jdevice))
	req_ += "&network=" + url.QueryEscape(string(jnetwork))
	if sgeo != "" {
		req_ += "&geo=" + url.QueryEscape(sgeo)
	}

	treq, err := http.NewRequest("GET", req_, nil)
	if err != nil {
		util.Log.Error("[gdt] http.NewRequest:%s", err.Error())
		return nil, err
	}
	treq.Header.Set("X-Forwarded-For", util.GetRealIp(r))
	treq.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	treq.Header.Set("Referer", r.Header.Get("Referer"))
	util.Log.Debug("[gdt] X-Forwarded-For:%s, User-Agent:%s, Referer:%s",
		util.GetRealIp(r), r.Header.Get("User-Agent"), r.Header.Get("Referer"))
	gres, err := dsp.Dspclient.Do(treq)
	if err != nil {
		return nil, err
	}
	defer gres.Body.Close()
	body, err := ioutil.ReadAll(gres.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (h *GdtHandler) BuildAdResponse(b []byte) (*adx.Response, error) {
	gres := &gdt.Response{}
	err := json.Unmarshal(b, gres)
	if err != nil {
		return nil, err
	}

	res := &adx.Response{
		Ret: gres.Ret,
		Msg: gres.Msg,
	}

	res.Data = make(map[string][]*adx.Ad)
	for k, v := range gres.Data {
		for _, gad := range v.List {
			ad := &adx.Ad{
				AdId:                     gad.AdId,
				ImpressionLink:           []string{gad.ImpressionLink},
				ClickLink:                gad.ClickLink,
				InteractType:             gad.InteractType,
				IsFullScreenInterstitial: gad.IsFullScreenInterstitial,
				HtmlSippet:               gad.HtmlSippet,
				CrtType:                  gad.CrtType,
				ImgUrl:                   []string{gad.ImgUrl},
				Img2Url:                  gad.Img2Url,
				Title:                    gad.Title,
				Description:              []string{gad.Description},
				SnapshotUrl:              gad.SnapshotUrl,
			}
			ad.AdTracking = make([]*adx.Tracking, 0)
			if "" != gad.VideoViewLink {
				atrack := &adx.Tracking{
					TrackingEvent: 101002,
					TrackingUrl:   []string{gad.VideoViewLink},
				}
				ad.AdTracking = append(ad.AdTracking, atrack)
			}
			if "" != gad.ConversionLink {
				event := []int32{102000, 102001, 102002}
				for _, e := range event {
					tk := &adx.Tracking{
						TrackingEvent: e,
						TrackingUrl:   []string{gad.ConversionLink},
					}
					ad.AdTracking = append(ad.AdTracking, tk)
				}
			}
			res.Data[k] = append(res.Data[k], ad)
		}
	}

	return res, nil
}
