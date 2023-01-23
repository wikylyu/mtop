package proxy

import (
	"net"
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

	buf := make([]byte, 4096)
	for {
		n, err := clientConn.Read(buf)
		if err != nil || n <= 0 {
			break
		}
		if _, err := clientConn.Write(buf[0:n]); err != nil {
			break
		}
	}
}
