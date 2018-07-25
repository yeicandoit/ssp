package baidu

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"ssp/dsp"
	"ssp/protocol/adx"
	"ssp/protocol/baidu"
	"ssp/util"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
)

const SidName = "baidu"

var deviceTypeMap = map[int32]baidu.Device_DeviceType{
	0: baidu.Device_PHONE, // Set unkown device as phone since baidu has no key for unkown device
	1: baidu.Device_PHONE,
	2: baidu.Device_TABLET,
}

var conTypeMap = map[int32]baidu.Network_ConnectionType{
	0: baidu.Network_CONNECTION_UNKNOWN,
	1: baidu.Network_WIFI,
	2: baidu.Network_CELL_2G,
	3: baidu.Network_CELL_3G,
	4: baidu.Network_CELL_4G,
}

var opeTypeMap = map[int32]baidu.Network_OperatorType{
	0: baidu.Network_UNKNOWN_OPERATOR,
	1: baidu.Network_CHINA_MOBILE,
	2: baidu.Network_CHINA_UNICOM,
	3: baidu.Network_CHINA_TELECOM,
}

var corTypeMap = map[int32]baidu.Gps_CoordinateType{
	1: baidu.Gps_WGS84,
	2: baidu.Gps_GCJ02,
	3: baidu.Gps_BD09,
}

var interMap = map[baidu.MaterialMeta_InteractionType]int32{
	baidu.MaterialMeta_SURFING:        0,
	baidu.MaterialMeta_DOWNLOAD:       1,
	baidu.MaterialMeta_DEEPLINK:       2,
	baidu.MaterialMeta_NO_INTERACTION: 100,
}

func init() {
	dsp.RegisterHandler(SidName, New())
}

type BaiduHandler struct {
	dsp.BaseHandler
}

func New() *BaiduHandler {
	h := &BaiduHandler{}
	h.Delegate = h
	h.Sid = SidName
	return h
}

func (h *BaiduHandler) SendDspRequest(r *http.Request, req *adx.Request) ([]byte, error) {
	baiduReq := &baidu.MobadsRequest{
		RequestId: proto.String(util.GenRandId()),
		ApiVersion: &baidu.Version{
			Major: proto.Uint32(5),
			Minor: proto.Uint32(2),
			Micro: proto.Uint32(0),
		},
		App: &baidu.App{
			AppId: func() *string {
				if nil == req.Media {
					return proto.String("")
				}
				return proto.String(req.Media.AppId)
			}(), //TODO Get app id from adx request or somewhere else
			ChannelId:  nil, //TODO Get channel id from adx request or somewhere else.
			AppVersion: nil, //TODO Get channel id from adx request or somewhere else.
			AppPackage: nil, //TODO Get channel id from adx request or somewhere else.
		},
	}
	dev := req.Device
	baiduReq.Device = &baidu.Device{
		DeviceType: func() *baidu.BidRequest_DeviceType {
			if dType, ok := deviceTypeMap[dev.DeviceType]; ok {
				return dType.Enum()
			}
			return baidu.Device_PHONE.Enum()
		}(),
		OsType: func() *baidu.Device_OsType {
			if "android" != dev.Os {
				return baidu.Device_IOS.Enum()
			}
			return baidu.Device_ANDROID.Enum()
		},
		OsVersion: func() *baidu.Version {
			v := &baidu.Version{}
			if "unknown" == dev.OsVersion {
				return v
			}
			s := strings.Split(dev.OsVersion, ".")
			vm := []**uint32{&v.Major, &v.Minor, &v.Micro}
			for index, value := range vm {
				if index < len(s) {
					ver, _ := strconv.Atoi(s[index])
				}
				*value = proto.Uint32(uint32(ver))
			}
			return v
		}(),
		Vendor: []byte(dev.Manufacturer),
		Model:  []byte(dev.Model),
		Udid: &baidu.UdId{
			Idfa:         proto.String(dev.Idfa),
			Imei:         proto.String(dev.Imei),
			ImeiMd5:      proto.String(dev.ImeiMd5),
			AndroidId:    proto.String(dev.AndroidId),
			AndroididMd5: proto.String(dev.AndroidIdMd5),
		},
		ScreenSize: &baidu.Size{
			Width:  proto.Uint32(dev.ScreenWidth),
			Height: proto.Uint32(dev.ScreenHeight),
		},
	}

	net := req.Network
	baiduReq.Network = &baidu.Network{
		Ipv4: util.GetRealIp(r),
		ConnectionType: func() *baidu.Network_ConnectionType {
			if conType, ok := conTypeMap[net.ConnectType]; ok {
				return conType.Enum()
			}
			return baidu.Network_CONNECTION_UNKNOWN.Enum()
		}(),
		OperatorType: func() *baidu.Network_OperatorType {
			if opeType, ok := opeTypeMap[net.Carrier]; ok {
				return opeType.Enum()
			}
			return baidu.Network_UNKNOWN_OPERATOR.Enum()
		}(),
	}

	geo := req.Geo
	baiduReq.Gps = &baidu.Gps{
		CoordinateType: func() *baidu.Gps_CoordinateType { //TODO adx proto doc should be updated
			if corType, ok := corTypeMap[geo.CoordinateType]; ok {
				return corType.Enum()
			}
			return baidu.Gps_GCJ02.Enum()
		}(),
		Longitude: proto.Float64(float64(geo.Lng) / 1000000),
		Latitude:  proto.Float64(float64(geo.Lat) / 1000000),
		Timestamp: proto.Uint32(uint32(geo.CoordTime)),
	}

	slot := req.Pos
	baiduReq.Adslot = &baidu.AdSlot{
		// AdslotId: proto.String(slot.Id), //TODO set slot id correctly
		// Video TODO should set?
		AdslotSize: &baidu.Size{
			Width:  proto.Uint32(uint32(slot.Width)),
			Height: proto.Uint32(uint32(slot.Height)),
		},
	}

	baiduReq_, _ = proto.Marshal(baiduReq)
	breq, err := http.NewRequest(util.KHttpPost, util.ServiceConfig.BaiduUrl, bytes.NewReader(baiduReq_))
	if err != nil {
		return nil, err
	}
	breq.Header.Set(util.KHttpContentType, util.KHttpContentTypeProto)

	bres, err := dsp.Dspclient.Do(breq)
	if err != nil {
		return nil, err
	}
	defer bres.Body.Close()
	body, err := ioutil.ReadAll(bres.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (h *BaiduHandler) BuildAdResponse(b []byte) (*adx.Response, error) {
	bres := &baidu.MobadsResponse{}
	err := proto.Unmarshal(b, bres)
	if err != nil {
		return nil, err
	}
	res := &adx.Response{
		Ret:  *bres.ErrorCode,
		Data: make(map[string]*adx.Ad),
	}

	tracking := &adx.Tracking{}

	for _, ad := range bres.Ads {
		adTracking := make([]*adx.Tracking, 0)
		for _, track := range ad.AdTracking {
			atr := &adx.Tracking{
				TrackingEvent: int32(*track.TrackingEvent),
				TrackingUrl:   track.TrackingUrl,
			}
			adTracking = adTracking.append(adTracking, atr)
		}
		adxAds := make([]*adx.Ad, 0)
		if nil != ad.MaterialMeta {
			ad.MetaGroup = append(ad.MetaGroup, ad.MaterialMeta)
		}
		for _, meta := range ad.MetaGroup {
			adxAd := &adx.Ad{
				AdId:           *ad.AdslotId,
				ImpressionLink: meta.WinNoticeUrl,
				//VideoViewLink: , TODO should consider how to merge gdt and baidu video monitor
				ClickLink: *meta.ClickUrl,
				InteractType: func() int32 {
					if interType, ok := interMap[*meta.InteractionType]; ok {
						return interType
					}
					return 0
				}(),
				AdTracking: adTracking,
				HtmlSippet: string(ad.HtmlSnippet),
				CrtType:    int32(*meta.CreativeType),
				ImgUrl:     meta.ImageSrc,
				Title:      *meta.AdTitle,
				Description: func() []string {
					desc := make([]string, 0)
					for _, d := range meta.Description {
						desc = append(desc, string(d))
					}
					return desc
				}(),
				VideoUrl: *meta.VideoUrl,
			}
			adxAds = append(adxAds, adxAd)
		}
		res.Data[*ad.AdslotId] = adxAds
	}

	return res, nil
}
