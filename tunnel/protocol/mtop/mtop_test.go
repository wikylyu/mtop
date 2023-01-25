package mtop

import (
	"bufio"
	"bytes"
	"net"
	"testing"
)

func testMTopAddress(t *testing.T, typ MTopAddressType, ip net.IP, domain string, port uint16, size int) {
	addr := NewMTopAddress(typ, ip, domain, port)
	addrBytes := addr.Bytes()
	if len(addrBytes) != size {
		t.Fatalf("invalid address bytes length: %d", len(addrBytes))
	}

	reader := bytes.NewReader(addrBytes)
	addr2, err := parseMTopAddress(bufio.NewReader(reader))
	if err != nil {
		t.Fatalf("parse address fail: %v", err)
	}

	if addr2.Type != typ {
		t.Fatalf("address type error: %v", addr2.Type)
	}
	if !addr2.IP.Equal(ip) {
		t.Fatalf("address ip error: %v", addr2.IP)
	}
	if addr2.Domain != domain {
		t.Fatalf("address domain error: %v", addr2.Domain)
	}
	if addr2.Port != port {
		t.Fatalf("address port error: %v", addr2.Port)
	}
}

func TestMTopAddress(t *testing.T) {
	testMTopAddress(t, MTopAddressTypeIPv4, net.IPv4(127, 0, 0, 1), "", 80, 7)
	testMTopAddress(t, MTopAddressTypeIPv6, net.IPv4(127, 0, 0, 1), "", 80, 19)
	testMTopAddress(t, MTopAddressTypeIPv4, net.IPv4(17, 4, 123, 4), "", 1180, 7)

	domain := "www.baidu.com"
	testMTopAddress(t, MTopAddressTypeDomain, nil, domain, 80, 4+len(domain))

	domain = "www.google.com"
	testMTopAddress(t, MTopAddressTypeDomain, nil, domain, 443, 4+len(domain))
}

func TestMTopAuthenticationMessage(t *testing.T) {
	username := "test111"
	password := "password"
	addr := NewMTopAddress(MTopAddressTypeIPv4, net.IPv4(127, 0, 0, 1), "", 80)
	m := NewMTopAuthenticationMessage(MTopVersion1, MTopMethodConnect, username, password, addr)
	mBytes := m.Bytes()
	if len(mBytes) != 4+len(username)+len(password)+7 {
		t.Fatalf("invalid message bytes length: %d", len(mBytes))
	}

	reader := bytes.NewReader(mBytes)
	m2, err := ParseMTopAuthenticationMessage(reader)
	if err != nil {
		t.Fatalf("parse message error: %v", err)
	}

	if m2.Version != MTopVersion1 {
		t.Fatalf("message version error: %v", m2.Version)
	}
	if m2.Method != MTopMethodConnect {
		t.Fatalf("message method error: %v", m2.Method)
	}
	if m2.Username != username {
		t.Fatalf("message username error: %v", m2.Username)
	}
	if m2.Password != password {
		t.Fatalf("message password error: %v", m2.Password)
	}
	if !bytes.Equal(m2.Address.Bytes(), addr.Bytes()) {
		t.Fatalf("message address error: %v", m2.Address)
	}
}
