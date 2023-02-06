package quic

import (
	"net"
	"time"

	quicgo "github.com/quic-go/quic-go"
)

/*
 * a simple wrapper to make quicgo.Stream compatible with net.Conn
 */
type Conn struct {
	s          quicgo.Stream
	remoteAddr net.Addr
}

func NewConn(s quicgo.Stream, remoteAddr net.Addr) *Conn {
	return &Conn{s: s, remoteAddr: remoteAddr}
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.s.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.s.Write(b)
}

func (c *Conn) Close() error {
	return c.s.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return nil
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.s.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.s.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.s.SetWriteDeadline(t)
}
