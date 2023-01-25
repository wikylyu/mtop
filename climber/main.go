package main

import (
	"bufio"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/elazarl/goproxy"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel"
	"github.com/wikylyu/mtop/tunnel/protocol/mtop"
)

const (
	AppName = "climber"
)

func init() {
	config.Init(AppName)
	config.InitLog()
	initServers()
}

type ServerConfig struct {
	Host     string `json:"host" yaml:"host"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	CA       string `json:"ca" yaml:"ca"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

var servers []*ServerConfig = nil

func initServers() {
	allServers := make([]*ServerConfig, 0)
	if err := config.Unmarshal("servers", &allServers); err != nil {
		panic(err)
	}
	servers = make([]*ServerConfig, 0)
	for _, s := range allServers {
		if s.Enabled {
			servers = append(servers, s)
		}
	}
}

/*  choose a server randomly */
func getServer() *ServerConfig {
	if len(servers) == 0 {
		return nil
	}
	i := rand.Intn(len(servers))
	return servers[i]
}

func runHTTPProxy() {
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
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		server := getServer()
		if server == nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "")
		}
		port, err := strconv.ParseUint(req.URL.Port(), 10, 16)
		if err != nil {
			port = 80
		}
		mc, err := mtop.Dial(server.CA, server.Host, server.Username, server.Password, req.URL.Hostname(), uint16(port))
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

		return req, resp
	})
	proxy.OnRequest().HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
		defer client.Close()
		server := getServer()
		if server == nil {
			return
		}
		port, err := strconv.ParseUint(req.URL.Port(), 10, 16)
		if err != nil {
			port = 80
		}
		mc, err := mtop.Dial(server.CA, server.Host, server.Username, server.Password, req.URL.Hostname(), uint16(port))
		if err != nil {
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

		tunnel.ConnForwarding(client, mc)
	})
	log.Infof("http proxy on %s", cfg.Listen)
	http.ListenAndServe(cfg.Listen, proxy)
}

func main() {

	go runHTTPProxy()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
