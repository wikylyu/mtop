package proxy

import (
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel/protocol/mtop"
)

type ServerConfig struct {
	Host     string `json:"host" yaml:"host"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Type     string `json:"type" yaml:"type"`
	CA       string `json:"ca" yaml:"ca"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

var servers []*ServerConfig = nil

func Init() {
	allServers := make([]*ServerConfig, 0)
	if err := config.Unmarshal("servers", &allServers); err != nil {
		panic(err)
	}
	servers = make([]*ServerConfig, 0)
	for _, s := range allServers {
		if s.Enabled {
			servers = append(servers, s)
		}
	}
}

/*  choose a server randomly */
func getServer() *ServerConfig {
	if len(servers) == 0 {
		return nil
	}
	i := rand.Intn(len(servers))
	return servers[i]
}

func DialHostAndPort(hostname string, port uint16) (*mtop.MTopClientConn, error) {
	server := getServer()
	if server == nil {
		return nil, errors.New("no available server")
	}
	var mc *mtop.MTopClientConn
	var err error
	if strings.ToLower(server.Type) == "quic" {
		mc, err = mtop.DialQUIC(server.CA, server.Host, server.Username, server.Password, hostname, port)
	} else {
		mc, err = mtop.DialTLS(server.CA, server.Host, server.Username, server.Password, hostname, port)
	}
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func DialURL(u *url.URL) (*mtop.MTopClientConn, error) {
	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		port = 80
	}
	return DialHostAndPort(u.Hostname(), uint16(port))
}
