package tunnel

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wikylyu/mtop/tunnel/protocol/mtop"
)

type Tunnel struct {
	clientConn *mtop.MTopServerConn
	remoteConn net.Conn
}

func NewTunnel(conn net.Conn) *Tunnel {
	return &Tunnel{
		clientConn: mtop.NewMTopServerConn(conn),
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

	if err := t.clientConn.Handshake(func(username, password, address string) bool {
		if username != "test" || password != "123456" {
			return false
		}
		remoteConn, err := net.Dial("tcp", address)
		if err != nil || remoteConn == nil {
			log.Debugf("dial %s error: %v", address, err)
			return false
		}
		t.remoteConn = remoteConn
		log.Debugf("connection established: %s <=> %s", clientConn.RemoteAddr().String(), address)
		return true
	}); err != nil {
		return
	}

	go func() {
		defer t.Close() // it's ok to close tunnel twice
		buf := make([]byte, 4096)
		for {
			t.remoteConn.SetReadDeadline(time.Now().Add(time.Minute * 2)) // 120 seconds for timeout
			n, err := t.remoteConn.Read(buf)
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

		if _, err := t.remoteConn.Write(buf[:n]); err != nil {
			break
		}
	}
}
