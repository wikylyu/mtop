package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel/protocol/mtop"
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
		mc, err := mtop.Dial(server.CA, server.Host, server.Username, server.Password, "www.baidu.com", 80)
		if err != nil {
			log.Warnf("mtop dial error: %v", err)
			continue
		}
		defer mc.Close()

		if n, err := mc.Write([]byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\nUser-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0\r\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\nAccept-Language: en,en-US;q=0.8,zh;q=0.5,zh-CN;q=0.3\r\nAccept-Encoding: gzip, deflate, br\r\nConnection: closed\r\n\r\n")); err != nil {
			log.Fatalf("write error: %v", err)
		} else {
			log.Debugf("write %d", n)
		}

		buf := make([]byte, 4096)
		for {
			n, err := mc.Read(buf)
			if err != nil || n <= 0 {
				break
			}
			fmt.Printf("%s", string(buf[:n]))
		}
	}

}
