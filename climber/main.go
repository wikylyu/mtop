package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel/protocol"
)

const (
	AppName = "climber"
)

func init() {
	config.Init(AppName)
	config.InitLog()
}

type ServerConfig struct {
	Host     string `json:"host" yaml:"host"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	CA       string `json:"ca" yaml:"ca"`
}

func main() {

	servers := make([]*ServerConfig, 0)
	if err := config.Unmarshal("servers", &servers); err != nil {
		panic(err)
	}
	for _, server := range servers {
		tlsConf := &tls.Config{}
		if server.CA != "" {
			certPool := x509.NewCertPool()
			pem, err := ioutil.ReadFile("../keys/ca.crt")
			if err != nil {
				panic(err)
			}
			if !certPool.AppendCertsFromPEM(pem) {
				panic("failed")
			}
			tlsConf.RootCAs = certPool
		}

		conn, err := tls.Dial("tcp", server.Host, tlsConf)
		if err != nil {
			log.Fatalf("connect to server error: %v", err)
		}
		defer conn.Close()

		req := protocol.NewMTopAuthenticationMessage(
			protocol.MTopVersion1, protocol.MTopMethodConnect, server.Username, server.Password,
			protocol.NewMTopAddress(protocol.MTopAddressTypeDomain, nil, "www.baidu.com", 80),
		)

		if _, err := conn.Write(req.Bytes()); err != nil {
			log.Fatalf("write error: %v", err)
		}

		resp, err := protocol.ParseMTopResponseMessage(conn)
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		log.Infof("verion: %d", resp.Version)
		log.Infof("Status: %d", resp.Status)

		if n, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\nUser-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0\r\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\nAccept-Language: en,en-US;q=0.8,zh;q=0.5,zh-CN;q=0.3\r\nAccept-Encoding: gzip, deflate, br\r\nConnection: closed\r\n\r\n")); err != nil {
			log.Fatalf("write error: %v", err)
		} else {
			log.Debugf("write %d", n)
		}

		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil || n <= 0 {
				break
			}
			fmt.Printf("%s", string(buf[:n]))
		}
	}

}
