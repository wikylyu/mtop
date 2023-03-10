package proxy

import (
	"bufio"
	"net"
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel"
)

type nilLogger struct {
}

func (l *nilLogger) Printf(format string, v ...interface{}) {

}

func RunHTTPProxy() {
	var cfg struct {
		Listen string `json:"listen" yaml:"listen"`
	}
	if err := config.Unmarshal("http", &cfg); err != nil {
		panic(err)
	} else if cfg.Listen == "" {
		return
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = log.GetLevel() > log.InfoLevel
	proxy.Logger = &nilLogger{}
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		mc, err := DialURL(req.URL)
		if err != nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "")
		}
		defer mc.Close()

		if err := req.Write(mc); err != nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "")
		}

		reader := bufio.NewReader(mc)
		resp, err := http.ReadResponse(reader, req)
		if err != nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "")
		}

		log.Infof("[http] %s %s %s", req.RemoteAddr, req.Method, req.URL.Hostname())

		return req, resp
	})
	proxy.OnRequest().HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
		defer client.Close()
		mc, err := DialURL(req.URL)
		if err != nil {
			log.Warnf("[http] Connect to %s error: %v", req.URL.String(), err)
			return
		}
		defer mc.Close()

		resp := http.Response{
			Status:     "200 Connection Established",
			StatusCode: http.StatusOK,
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Header:     nil,
			Body:       nil,
		}
		if err := resp.Write(client); err != nil {
			return
		}

		log.Infof("[http] Connection established %s - %s", req.RemoteAddr, req.URL.Hostname())
		tunnel.Transmit(client, mc)
		log.Infof("[http] Connection closed %s - %s", req.RemoteAddr, req.URL.Hostname())
	})
	log.Infof("starting http proxy on %s", cfg.Listen)
	if err := http.ListenAndServe(cfg.Listen, proxy); err != nil {
		log.Errorf("http server error: %v", err)
	}
}
