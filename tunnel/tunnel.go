package tunnel

import (
	"io"
	"net"
	"time"

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

	/* read authenticantion message from client */
	req, err := protocol.ParseMTopAuthenticationMessage(clientConn)
	if err != nil {
		if err != io.EOF {
			log.Warnf("parse mtop request error: %v", err)
		}
		return
	}
	/* verify message */
	if req.Version != protocol.MTopVersion1 {
		log.Warnf("unsupported version: %v", req.Version)
		return
	}
	if req.Method != protocol.MTopMethodConnect {
		log.Warnf("invalid mtop method: %v", req.Method)
		return
	}
	if req.Username != "test" || req.Password != "123456" {
		log.Debugf("invalid auth: %s:%s", req.Username, req.Password)
		return
	}

	// log.Infof("version: %d", req.Version)
	// log.Infof("username: %s", req.Username)
	// log.Infof("password: %s", req.Password)
	// log.Infof("host: %s:%d", req.Address.Domain, req.Address.Port)

	/* connect to remote server */
	remoteConn, err := net.Dial("tcp", req.Address.String())
	if err != nil || remoteConn == nil {
		log.Debugf("dial %s error: %v", req.Address.String(), err)
		return
	}
	t.remoteConn = remoteConn

	log.Debugf("connection established: %s <=> %s", clientConn.RemoteAddr().String(), req.Address.String())

	/*
	 * response with success message
	 */
	resp := protocol.NewMTopResponseMessage(protocol.MTopVersion1, protocol.MTopResponseStatusSuccess)
	if _, err := clientConn.Write(resp.Bytes()); err != nil {
		return
	}

	go func() {
		defer t.Close() // it's ok to close tunnel twice
		buf := make([]byte, 4096)
		for {
			remoteConn.SetReadDeadline(time.Now().Add(time.Minute * 2)) // 120 seconds for timeout
			n, err := remoteConn.Read(buf)
			if err != nil || n <= 0 {
				break
			}

			if _, err := clientConn.Write(buf[:n]); err != nil {
				break
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		clientConn.SetReadDeadline(time.Now().Add(time.Minute * 2)) // 120 seconds for timeout
		n, err := clientConn.Read(buf)
		if err != nil || n <= 0 {
			break
		}

		if _, err := remoteConn.Write(buf[:n]); err != nil {
			break
		}
	}
	log.Debugf("connection closed: %s <=> %s", clientConn.RemoteAddr().String(), req.Address.String())
}
