package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"tripod/devkit"
)

func GenRandId() string {
	s := fmt.Sprintf("%d%s", time.Now().Unix(), devkit.RandString(6))
	return devkit.HashFunc(s)
}

//build safe info
func BuildSafeUrl(info string) string {
	return url.QueryEscape(base64.URLEncoding.EncodeToString([]byte(info)))
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
