package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/tunnel"
)

const (
	AppName = "mtop"
)

func init() {
	config.Init(AppName)
	initLog()
}

func initLog() {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
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
