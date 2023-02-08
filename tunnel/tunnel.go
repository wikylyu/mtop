package tunnel

import (
	"io"
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
			log.Debugf("User %s password incorrect: %s", password)
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
		for {
			c1.SetReadDeadline(time.Now().Add(time.Minute)) // 60 seconds for timeout
			if _, err := io.Copy(c2, c1); err != nil {
				break
			}
		}
	}()

	for {
		c2.SetReadDeadline(time.Now().Add(time.Minute)) // 60 seconds for timeout
		if _, err := io.Copy(c1, c2); err != nil {
			break
		}
	}
}
