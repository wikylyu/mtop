package proxy

import (
	"context"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	socks "github.com/wikylyu/gosocks"
	"github.com/wikylyu/mtop/config"
)

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
		return nil, socks.NewError(socks.RequestReplyHostUnreachable, err)
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

func RunSOCKS5Proxy() {
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

	p := socks.NewProxy(cfg.Listen, handler, nil, time.Duration(cfg.Timeout)*time.Second, log.New())

	log.Infof("starting SOCKS server on %s", cfg.Listen)
	if err := p.Start(); err != nil {
		log.Warnf("socks5 server error: %v", err)
	}
	<-p.Done
}
