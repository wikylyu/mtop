package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/elazarl/goproxy"
	socks "github.com/firefart/gosocks"
	log "github.com/sirupsen/logrus"
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

func DialHostAndPort(hostname string, port uint16) (*mtop.MTopClientConn, error) {
	server := getServer()
	if server == nil {
		return nil, errors.New("no available server")
	}
	mc, err := mtop.Dial(server.CA, server.Host, server.Username, server.Password, hostname, port)
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func DialURL(u *url.URL) (*mtop.MTopClientConn, error) {
	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		port = 80
	}
	return DialHostAndPort(u.Hostname(), uint16(port))
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

		return req, resp
	})
	proxy.OnRequest().HijackConnect(func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
		defer client.Close()
		mc, err := DialURL(req.URL)
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

		tunnel.Transmit(client, mc)
	})
	log.Infof("starting http proxy on %s", cfg.Listen)
	http.ListenAndServe(cfg.Listen, proxy)
}

type socks5Handler struct {
}

func (s *socks5Handler) PreHandler(req socks.Request) (io.ReadWriteCloser, *socks.Error) {
	var addr string
	if req.AddressType == socks.RequestAddressTypeDomainname {
		addr = string(req.DestinationAddress)
	} else {
		addr = net.IP(req.DestinationAddress).String()
	}
	mc, err := DialHostAndPort(addr, req.DestinationPort)
	if err != nil {
		return nil, &socks.Error{Reason: socks.RequestReplyHostUnreachable}
	}
	return mc, nil
}

func (s *socks5Handler) Refresh(ctx context.Context) {
}

func (s *socks5Handler) CopyFromRemoteToClient(ctx context.Context, remote io.ReadCloser, client io.WriteCloser) error {
	_, err := io.Copy(client, remote)
	return err
}

func (s *socks5Handler) CopyFromClientToRemote(ctx context.Context, client io.ReadCloser, remote io.WriteCloser) error {
	_, err := io.Copy(remote, client)
	return err
}

func (s *socks5Handler) Cleanup() error {
	return nil
}

func runSOCKS5Proxy() {
	var cfg struct {
		Listen  string `json:"listen" yaml:"listen"`
		Timeout int    `json:"timeout" yaml:"timeout"`
	}
	if err := config.Unmarshal("socks5", &cfg); err != nil {
		panic(err)
	} else if cfg.Listen == "" {
		return
	}

	handler := &socks5Handler{}

	p := socks.Proxy{
		ServerAddr:   cfg.Listen,
		Proxyhandler: handler,
		Timeout:      time.Duration(cfg.Timeout) * time.Second,
		Log:          log.New(),
	}
	log.Infof("starting SOCKS server on %s", cfg.Listen)
	if err := p.Start(); err != nil {
		log.Warnf("socks5 server error: %v", err)
	}
	<-p.Done
}

func main() {
	go runHTTPProxy()
	go runSOCKS5Proxy()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
