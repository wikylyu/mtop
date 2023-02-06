package main

import (
	"context"
	"crypto/tls"
	"flag"
	"strings"

	quicgo "github.com/quic-go/quic-go"
	log "github.com/sirupsen/logrus"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
	"github.com/wikylyu/mtop/tunnel"
	"github.com/wikylyu/mtop/tunnel/protocol/quic"
)

const (
	AppName = "mtop"
)

func initMTop(configFile string) {
	config.Init(AppName, configFile)
	config.InitLog()
	initDatabase()
}

func initDatabase() {
	var cfg struct {
		Debug      bool   `json:"debug" yaml:"debug"`
		DriverName string `json:"driverName" yaml:"driverName"`
		DSN        string `json:"dsn" yaml:"dsn"`
	}
	if err := config.Unmarshal("db", &cfg); err != nil {
		panic(err)
	}
	if err := db.Init(cfg.DriverName, cfg.DSN, cfg.Debug); err != nil {
		panic(err)
	}
}

func initServerConfig() (string, string, *tls.Config) {
	var cfg struct {
		Type   string `json:"type" yaml:"type"`
		Listen string `json:"listen" yaml:"listen"`
		CRT    string `json:"crt" yaml:"crt"`
		Key    string `json:"key" yaml:"key"`
	}
	if err := config.Unmarshal("server", &cfg); err != nil {
		panic(err)
	}
	if cfg.CRT == "" || cfg.Key == "" {
		log.Fatalf("certificate key not configured")
	}

	cer, err := tls.LoadX509KeyPair(cfg.CRT, cfg.Key)
	if err != nil {
		log.Fatalf("load x509 key error: %v", err)
	}

	return strings.ToLower(cfg.Type), cfg.Listen, &tls.Config{
		Certificates: []tls.Certificate{cer},
		NextProtos:   []string{"mtop"},
	}
}

func runQUICServer(listen string, config *tls.Config) {
	listener, err := quicgo.ListenAddr(listen, config, nil)
	if err != nil {
		log.Fatalf("listen on %s error: %v", listen, err)
	}
	log.Infof("[QUIC] listen on %s", listen)

	defer listener.Close()
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Errorf("[QUIC] accept error:%v", err)
			continue
		}
		go func() {
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				log.Warnf("[QUIC] accept stream error:%v", err)
				return
			}
			t := tunnel.NewTunnel(quic.NewConn(stream, conn.RemoteAddr()))
			t.Run()
		}()
	}
}

func runTLSServer(listen string, config *tls.Config) {
	listener, err := tls.Listen("tcp", listen, config)

	if err != nil {
		log.Fatalf("[TLS] listen on %s error: %v", listen, err)
	}
	log.Infof("[TLS] listen on %s", listen)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Warnf("[TLS] accept error: %v", err)
			continue
		}
		t := tunnel.NewTunnel(conn)
		go t.Run()
	}
}

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.Parse()

	initMTop(configFile)

	stype, listen, config := initServerConfig()
	if stype == "quic" {
		runQUICServer(listen, config)
	} else {
		runTLSServer(listen, config)
	}
}
