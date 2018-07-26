package dsp

import (
	"net/http"
	"ssp/protocol/adx"
	"ssp/util"
)

var Dspclient = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   1000 * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost: 100,
	},
	Timeout: 1000 * time.Millisecond,
}

var HandlerMap = map[string]HandlerDelegate{}

type HandlerDelegate interface {
	SendDspRequest(r *http.Request, req *adx.Request) ([]byte, error)
	BuildAdResponse(b []byte) (*adx.Response, error)
}

type BaseHandler struct {
	Delegate HandlerDelegate
	Sid      string
}

func RegisterHandler(name string, handler HandlerDelegate) {
	if _, found := HandlerMap[name]; found {
		panic("handler already exists. name:" + name)
	}
	HandlerMap[name] = handler
	util.Log.Info("Register handler %s", name)
}

func (self *BaseHandler) VerifyRequest(req *adx.Request) error {
	if nil == req {
		return errors.Error("adx.Request is nil")
	}

	if nil == req.Device {
		return errors.Error("Device of adx.Request is nil")
	}

	if nil == req.Network {
		return errors.Error("Network of adx.Request is nil")
	}

	if nil == req.Pos {
		return errors.Error("Pos of adx.Request is nil")
	}

	return nil
}

func (self *BaseHandler) Handle(r *http.Request, req *adx.Request) (*adx.Response, error) {
	if err := self.VerifyRequest(req); err != nil {
		return nil, err
	}
	dspRes, err := self.Delegate.SendDspRequest(r, req)
	if err != nil {
		return nil, err
	}
	res, err := self.Delegate.BuildAdResponse(dspRes)
	if err != nil {
		return nil, err
	}
	return res, nil
}
