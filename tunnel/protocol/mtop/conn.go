package mtop

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"

	quicgo "github.com/quic-go/quic-go"
	"github.com/wikylyu/mtop/tunnel/protocol/quic"
)

type MTopServerConn struct {
	c   net.Conn
	req *MTopAuthenticationMessage
}

func NewMTopServerConn(c net.Conn) *MTopServerConn {
	return &MTopServerConn{
		c: c,
	}
}

func (mc *MTopServerConn) Handshake(auth func(username, password, address string) bool) error {
	req, err := ParseMTopAuthenticationMessage(mc.c)
	if err != nil {
		return err
	}
	mc.req = req
	/* verify message */
	if req.Version != MTopVersion1 {
		return ErrInvalidVersion
	}
	if req.Method != MTopMethodConnect {
		return ErrInvalidMethod
	}
	if !auth(req.Username, req.Password, req.Address.String()) {
		return ErrInvalidMessage
	}
	resp := NewMTopResponseMessage(MTopVersion1, MTopResponseStatusSuccess)
	if _, err := mc.Write(resp.Bytes()); err != nil {
		return err
	}
	return nil
}

func (mc *MTopServerConn) Read(b []byte) (int, error) {
	return mc.c.Read(b)
}

func (mc *MTopServerConn) Write(b []byte) (int, error) {
	return mc.c.Write(b)
}

func (mc *MTopServerConn) Close() error {
	return mc.c.Close()
}

func (mc *MTopServerConn) LocalAddr() net.Addr {
	return mc.c.LocalAddr()
}

func (mc *MTopServerConn) RemoteAddr() net.Addr {
	return mc.c.RemoteAddr()
}

func (mc *MTopServerConn) SetDeadline(t time.Time) error {
	return mc.c.SetDeadline(t)
}

func (mc *MTopServerConn) SetReadDeadline(t time.Time) error {
	return mc.c.SetReadDeadline(t)
}

func (mc *MTopServerConn) SetWriteDeadline(t time.Time) error {
	return mc.c.SetWriteDeadline(t)
}

type MTopClientConn struct {
	c   net.Conn
	req *MTopAuthenticationMessage
}

func NewMTopClientConn(c net.Conn, username, password string, addr *MTopAddress) *MTopClientConn {
	return &MTopClientConn{
		c:   c,
		req: NewMTopAuthenticationMessage(MTopVersion1, MTopMethodConnect, username, password, addr),
	}
}

func (mc *MTopClientConn) Connect() error {
	if _, err := mc.c.Write(mc.req.Bytes()); err != nil {
		return err
	}

	resp, err := ParseMTopResponseMessage(mc.c)
	if err != nil {
		return err
	} else if resp.Status != MTopResponseStatusSuccess {
		return ErrInvalidMessage
	}
	return nil
}

func (mc *MTopClientConn) Read(b []byte) (int, error) {
	return mc.c.Read(b)
}

func (mc *MTopClientConn) Write(b []byte) (int, error) {
	return mc.c.Write(b)
}

func (mc *MTopClientConn) Close() error {
	return mc.c.Close()
}

func (mc *MTopClientConn) LocalAddr() net.Addr {
	return mc.c.LocalAddr()
}

func (mc *MTopClientConn) RemoteAddr() net.Addr {
	return mc.c.RemoteAddr()
}

func (mc *MTopClientConn) SetDeadline(t time.Time) error {
	return mc.c.SetDeadline(t)
}

func (mc *MTopClientConn) SetReadDeadline(t time.Time) error {
	return mc.c.SetReadDeadline(t)
}

func (mc *MTopClientConn) SetWriteDeadline(t time.Time) error {
	return mc.c.SetWriteDeadline(t)
}

func getTLSConfig(host, ca, proto string) (*tls.Config, error) {
	serverName, _, _ := net.SplitHostPort(host)
	tlsConf := &tls.Config{
		NextProtos:         []string{proto},
		InsecureSkipVerify: false,
		ServerName:         serverName,
	}
	if ca != "" {
		certPool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, errors.New("append certs from pem failed")
		}
		tlsConf.RootCAs = certPool
	}
	return tlsConf, nil
}

func parseMTopAddressFromHost(target string, port uint16) *MTopAddress {
	var addr *MTopAddress = nil
	ip := net.ParseIP(target)
	if ip == nil {
		addr = NewMTopAddress(MTopAddressTypeDomain, nil, target, port)
	} else if ip.To4() == nil {
		addr = NewMTopAddress(MTopAddressTypeIPv6, ip, "", port)
	} else {
		addr = NewMTopAddress(MTopAddressTypeIPv4, ip.To4(), "", port)
	}
	return addr
}

func DialTLS(ca, server string, username, password string, target string, port uint16, proto string) (*MTopClientConn, error) {
	addr := parseMTopAddressFromHost(target, port)

	tlsConf, err := getTLSConfig(server, ca, proto)
	if err != nil {
		return nil, err
	}

	conn, err := tls.Dial("tcp", server, tlsConf)
	if err != nil {
		return nil, err
	}
	mc := NewMTopClientConn(conn, username, password, addr)
	if err := mc.Connect(); err != nil {
		return nil, err
	}
	return mc, nil
}

var quicConns map[string][]quicgo.Connection = make(map[string][]quicgo.Connection)
var quicMutex sync.Mutex

func getQuicConn(server string) quicgo.Connection {
	defer quicMutex.Unlock()
	quicMutex.Lock()
	conns := quicConns[server]
	if len(conns) > 0 {
		conn := conns[0]
		quicConns[server] = conns[1:]
		return conn
	}
	return nil
}

func setQuicConn(server string, conn quicgo.Connection) {
	defer quicMutex.Unlock()
	quicMutex.Lock()
	conns := quicConns[server]
	conns = append(conns, conn)
	quicConns[server] = conns
}

func DialQUIC(ca, server string, username, password string, target string, port uint16, proto string) (*MTopClientConn, error) {
	addr := parseMTopAddressFromHost(target, port)

	tlsConf, err := getTLSConfig(server, ca, proto)
	if err != nil {
		return nil, err
	}
	var stream quicgo.Stream = nil
	var mc *MTopClientConn = nil

	conn := getQuicConn(server)
	if conn != nil {
		stream, _ = conn.OpenStreamSync(context.Background())
		if stream != nil {
			mc = NewMTopClientConn(quic.NewConn(stream, conn.RemoteAddr()), username, password, addr)
			if err := mc.Connect(); err == nil {
				setQuicConn(server, conn)
				return mc, nil
			}
		}
	}

	conn, err = quicgo.DialAddr(context.Background(), server, tlsConf, &quicgo.Config{
		HandshakeIdleTimeout: time.Second * 5,
		MaxIdleTimeout:       time.Second * 10,
	})

	if err != nil {
		return nil, err
	}

	stream, err = conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	mc = NewMTopClientConn(quic.NewConn(stream, conn.RemoteAddr()), username, password, addr)
	if err := mc.Connect(); err != nil {
		return nil, err
	}
	setQuicConn(server, conn)
	return mc, nil
}
