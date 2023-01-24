package tunnel

import (
	"io"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/wikylyu/mtop/tunnel/protocol"
)

type Tunnel struct {
	clientConn net.Conn
	remoteConn net.Conn
}

func NewTunnel(conn net.Conn) *Tunnel {
	return &Tunnel{
		clientConn: conn,
		remoteConn: nil,
	}
}

func (t *Tunnel) Close() {
	if t.clientConn != nil {
		t.clientConn.Close()
	}
	if t.remoteConn != nil {
		t.remoteConn.Close()
	}
}

func (t *Tunnel) Run() {
	defer t.Close()

	clientConn := t.clientConn

	for {
		reqMessage, err := protocol.ParseMTopAuthenticationMessage(clientConn)
		if err != nil {
			if err != io.EOF {
				log.Warnf("parse mtop request error: %v", err)
			}
			break
		}

		log.Infof("version: %d", reqMessage.Version)
		log.Infof("username: %s", reqMessage.Username)
		log.Infof("password: %s", reqMessage.Password)
		log.Infof("host: %s:%d", reqMessage.Address.Domain, reqMessage.Address.Port)

		if _, err := clientConn.Write(reqMessage.Bytes()); err != nil {
			break
		}
	}
}
