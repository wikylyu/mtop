package tunnel

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wikylyu/mtop/db"
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

	var remoteAddress string

	if err := t.clientConn.Handshake(func(username, password, address string) bool {
		user, err := db.GetUserByUsername(username)
		if err != nil {
			log.Errorf("GetUserByUsername() error: %v", err)
			return false
		} else if user == nil {
			log.Debugf("User %s not found", username)
			return false
		} else if !user.Auth(password) {
			log.Debugf("User %s password incorrect: %s", username, password)
			return false
		}

		remoteConn, err := net.Dial("tcp", address)
		if err != nil || remoteConn == nil {
			log.Debugf("dial %s error: %v", address, err)
			return false
		}
		t.remoteConn = remoteConn
		remoteAddress = address
		return true
	}); err != nil {
		return
	}
	log.Debugf("connection established: %s - %s", clientConn.RemoteAddr().String(), remoteAddress)

	Transmit(t.clientConn, t.remoteConn)

	log.Debugf("connection closed: %s - %s", clientConn.RemoteAddr().String(), remoteAddress)
}

func Transmit(c1, c2 net.Conn) {
	defer c1.Close() // it's ok to close connection twice
	defer c2.Close()
	go func() {
		defer c1.Close() // it's ok to close connection twice
		defer c2.Close()
		buf := make([]byte, 4096)
		for {
			c1.SetReadDeadline(time.Now().Add(time.Minute)) // 60 seconds for timeout
			n, err := c1.Read(buf)
			if err != nil || n <= 0 {
				break
			}
			if _, err := c2.Write(buf[:n]); err != nil {
				break
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		c2.SetReadDeadline(time.Now().Add(time.Minute)) // 60 seconds for timeout
		n, err := c2.Read(buf)
		if err != nil || n <= 0 {
			break
		}
		if _, err := c1.Write(buf[:n]); err != nil {
			break
		}
	}
}
