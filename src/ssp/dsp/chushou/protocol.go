package chushou

type Pos struct {
	Id                            int64  `json:"id"`
	Width                         int32  `json:"width"`
	Height                        int32  `json:"height"`
	SupportFullScreenInterstitial bool   `json:"support_full_screen_interstitial,omitempty"`
	AdCount                       int32  `json:"ad_count"`
	NeedRenderedAd                bool   `json:"need_rendered_ad,omitempty"`
	LastAdIds                     string `json:"last_ad_ids,omitempty"`
	Channel                       int32  `json:"channel,omitempty"`
	PageNumber                    int32  `json:"page_number,omitempty"`
}

type Media struct {
	AppId       string `json:"app_id"`
	AppBunbleId string `json:"app_bundle_id"`
}

type Device struct {
	Os           string `json:"os"`
	OsVersion    string `json:"os_version"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	DeviceType   int32  `json:"device_type"`
	ScreenWidth  int32  `json:"screen_width,omitempty"`
	ScreenHeight int32  `json:"screen_height,omitempty"`
	Dpi          int32  `json:"dpi,omitempty"`
	Orientation  int32  `json:"orientation,omitempty"`
	Idfa         string `json:"idfa"`
	Imei         string `json:"imei"`
	ImeiMd5      string `json:"imei_md5"`
	AndroidId    string `json:"android_id"`
	AndroidIdMd5 string `json:"android_id_md5"`
	AndroidAdId  string `json:"android_ad_id,omitempty"`
}

type Network struct {
	ConnectType int32 `json:"connect_type"`
	Carrier     int32 `json:"carrier"`
}

type Geo struct {
	Lat              int32   `json:"connect_type,omitempty"`
	Lng              int32   `json:"carrier,omitempty"`
	LocationAccuracy float64 `json:"location_accuracy,omitempty"`
	CoordTime        int64   `json:"coord_time,omitempty"`
}

//responsepart
type GdtResponse struct {
	Ret  int64    `json:"ret"`
	Msg  string   `json:"msg"`
	Data *GdtData `json:"data"`
}

type GdtData struct {
	PosId *GdtPos
}

type GdtPos struct {
	List []*GdtAd `json:"list"`
}

type GdtAd struct {
	Type                     string   `json:"type,omitempty"`
	AdId                     string   `json:"ad_id"`
	ImpressionLink           string   `json:"impression_link"`
	VideoViewLink            string   `json:"video_view_link,omitempty"`
	ClickLink                string   `json:"click_link"`
	InteractType             int32    `json:"interact_type"`
	ConversionLink           string   `json:"conversion_link,omitempty"`
	IsFullScreenInterstitial bool     `json:"is_full_screen_interstitial,omitempty"`
	HtmlSippet               string   `json:"html_snippet,omitempty"`
	CrtType                  int32    `json:"crt_type,omitempty"`
	ImgUrl                   string   `json:"img_url,omitempty"`
	Img2Url                  string   `json:"img2d_url,omitempty"`
	Title                    string   `json:"title,omitempty"`
	Description              string   `json:"description,omitempty"`
	SnapshotUrl              []string `json:"snapshot_url,omitempty"`
	VideoUrl                 string   `json:"video_url,omitempty"`
}

type GdtInformation struct {
	Type            string   `json:"type,omitempty"`
	InformationType string   `json:"information_type"`
	From            string   `json:"from"`
	Title           string   `json:"title"`
	Images          []string `json:"images"`
	IsBigPic        int32    `json:"is_big_pic"`
	Url             string   `json:"url"`
	CommentCount    int32    `json:"comment_count,omitempty"`
	PlayCount       int32    `json:"play_count,omitempty"`
	RunTime         int32    `json:"run_time,omitempty"`
}
