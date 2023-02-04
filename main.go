package main

import (
	"crypto/tls"
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
	"github.com/wikylyu/mtop/tunnel"
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

func initTSLConfig() (string, *tls.Config) {
	var cfg struct {
		Listen string `json:"listen" yaml:"listen"`
		CRT    string `json:"crt" yaml:"crt"`
		Key    string `json:"key" yaml:"key"`
	}
	if err := config.Unmarshal("tls", &cfg); err != nil {
		panic(err)
	}
	if cfg.CRT == "" || cfg.Key == "" {
		log.Fatalf("certificate key not configured")
	}

	cer, err := tls.LoadX509KeyPair(cfg.CRT, cfg.Key)
	if err != nil {
		log.Fatalf("load x509 key error: %v", err)
	}

	return cfg.Listen, &tls.Config{Certificates: []tls.Certificate{cer}}
}

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.Parse()

	initMTop(configFile)

	listen, config := initTSLConfig()
	ln, err := tls.Listen("tcp", listen, config)
	if err != nil {
		log.Fatalf("listen on %s error: %v", listen, err)
	}
	log.Infof("listen on %s", listen)
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
