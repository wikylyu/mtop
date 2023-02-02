package main

import (
	"crypto/tls"

	log "github.com/sirupsen/logrus"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
	"github.com/wikylyu/mtop/tunnel"
)

const (
	AppName = "mtop"
)

func init() {
	config.Init(AppName)
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

func main() {
	var tlsCfg struct {
		Listen string `json:"listen" yaml:"listen"`
		CRT    string `json:"crt" yaml:"crt"`
		Key    string `json:"key" yaml:"key"`
	}
	if err := config.Unmarshal("tls", &tlsCfg); err != nil {
		panic(err)
	}

	cer, err := tls.LoadX509KeyPair(tlsCfg.CRT, tlsCfg.Key)
	if err != nil {
		log.Fatalf("load x509 key error: %v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", tlsCfg.Listen, config)
	if err != nil {
		log.Fatalf("listen on %s error: %v", tlsCfg.Listen, err)
	}
	log.Infof("listen on %s", tlsCfg.Listen)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("accept error: %v", err)
			continue
		}
		t := tunnel.NewTunnel(conn)
		go t.Run()
	}
}
