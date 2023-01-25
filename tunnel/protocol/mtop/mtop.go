package mtop

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	MTopVersion1 = 0x1
)

const (
	MTopMethodConnect = 0x1
)

const (
	MTopAddressTypeIPv4   = 0x1
	MTopAddressTypeDomain = 0x3
	MTopAddressTypeIPv6   = 0x4
)

var (
	ErrInvalidVersion = errors.New("invalid version")
	ErrInvalidMethod  = errors.New("invalid method")
	ErrInvalidMessage = errors.New("invalid message")
)

type MTopAddressType byte

type MTopAddress struct {
	Type   MTopAddressType
	IP     net.IP
	Domain string
	Port   uint16
}

func NewMTopAddress(t MTopAddressType, ip net.IP, domain string, port uint16) *MTopAddress {
	return &MTopAddress{
		Type:   t,
		IP:     ip,
		Domain: domain,
		Port:   port,
	}
}

func parseMTopAddress(reader *bufio.Reader) (*MTopAddress, error) {
	t, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	var ip net.IP
	var domain string
	var port uint16
	if t == MTopAddressTypeIPv4 || t == MTopAddressTypeIPv6 {
		len := 4
		if t == MTopAddressTypeIPv6 {
			len = 16
		}
		buf := make([]byte, len)
		if n, err := io.ReadFull(reader, buf); err != nil {
			return nil, err
		} else if n != len {
			return nil, ErrInvalidMessage
		}
		ip = net.IP(buf)
	} else if t == MTopAddressTypeDomain {
		domain, err = parseText(reader)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, ErrInvalidMessage
	}
	if err := binary.Read(reader, binary.BigEndian, &port); err != nil {
		return nil, err
	}
	return NewMTopAddress(MTopAddressType(t), ip, domain, port), nil
}

/*
 * return bytes used for packaging
 */
func (address *MTopAddress) Bytes() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, address.Type)
	if address.Type == MTopAddressTypeDomain {
		binary.Write(&buf, binary.BigEndian, byte(len(address.Domain)))
		binary.Write(&buf, binary.BigEndian, []byte(address.Domain))
	} else if address.Type == MTopAddressTypeIPv4 {
		binary.Write(&buf, binary.BigEndian, []byte(address.IP.To4()))
	} else {
		binary.Write(&buf, binary.BigEndian, []byte(address.IP.To16()))
	}
	binary.Write(&buf, binary.BigEndian, address.Port)
	return buf.Bytes()
}

/*
 * format address like hostname:port
 */
func (address *MTopAddress) String() string {
	if address.Type == MTopAddressTypeDomain {
		return fmt.Sprintf("%s:%d", address.Domain, address.Port)
	}
	return fmt.Sprintf("%s:%d", address.IP.String(), address.Port)
}

type MTopAuthenticationMessage struct {
	Version  byte
	Method   byte
	Username string
	Password string
	Address  *MTopAddress
}

func NewMTopAuthenticationMessage(ver, method byte, username, password string, address *MTopAddress) *MTopAuthenticationMessage {
	return &MTopAuthenticationMessage{
		Version:  ver,
		Method:   method,
		Username: username,
		Password: password,
		Address:  address,
	}
}

func ParseMTopAuthenticationMessage(c io.Reader) (*MTopAuthenticationMessage, error) {
	reader := bufio.NewReader(c)

	ver, err := reader.ReadByte()
	if err != nil {
		return nil, err
	} else if ver < MTopVersion1 || ver > MTopVersion1 {
		return nil, ErrInvalidVersion
	}
	method, err := reader.ReadByte()
	if err != nil {
		return nil, err
	} else if method != MTopMethodConnect {
		return nil, ErrInvalidMethod
	}
	username, err := parseText(reader)
	if err != nil {
		return nil, err
	}
	password, err := parseText(reader)
	if err != nil {
		return nil, err
	}
	address, err := parseMTopAddress(reader)
	if err != nil {
		return nil, err
	}
	if reader.Buffered() > 0 { /* Extra data is not allowed */
		return nil, ErrInvalidMessage
	}

	return NewMTopAuthenticationMessage(ver, method, username, password, address), nil
}

func parseText(reader *bufio.Reader) (string, error) {
	len, err := reader.ReadByte()
	if err != nil {
		return "", err
	}
	buf := make([]byte, len)
	n, err := io.ReadFull(reader, buf)
	if err != nil {
		return "", err
	} else if n != int(len) {
		return "", ErrInvalidMessage
	}
	return string(buf), nil
}

func (m *MTopAuthenticationMessage) Bytes() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, m.Version)
	binary.Write(&buf, binary.BigEndian, m.Method)
	binary.Write(&buf, binary.BigEndian, byte(len(m.Username)))
	binary.Write(&buf, binary.BigEndian, []byte(m.Username))
	binary.Write(&buf, binary.BigEndian, byte(len(m.Password)))
	binary.Write(&buf, binary.BigEndian, []byte(m.Password))

	binary.Write(&buf, binary.BigEndian, m.Address.Bytes())
	return buf.Bytes()
}

const (
	MTopResponseStatusSuccess            = 0x0
	MTopResponseStatusUnsupportedVersion = 0x1
	MTopResponseStatusAuthFailure        = 0x2
	MTopResponseStatusConnectionFailure  = 0x3
)

type MTopResponseMessage struct {
	Version byte
	Status  byte
}

func NewMTopResponseMessage(ver, status byte) *MTopResponseMessage {
	return &MTopResponseMessage{
		Version: ver,
		Status:  status,
	}
}

func ParseMTopResponseMessage(c net.Conn) (*MTopResponseMessage, error) {
	reader := bufio.NewReader(c)
	ver, err := reader.ReadByte()
	if err != nil {
		return nil, err
	} else if ver < MTopVersion1 || ver > MTopVersion1 {
		return nil, ErrInvalidVersion
	}
	status, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if reader.Buffered() > 0 {
		return nil, ErrInvalidMessage
	}

	return NewMTopResponseMessage(ver, status), nil
}

func (m *MTopResponseMessage) Bytes() []byte {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, m.Version)
	binary.Write(&buf, binary.BigEndian, m.Status)
	return buf.Bytes()
}
