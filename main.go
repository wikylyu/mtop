package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	quicgo "github.com/quic-go/quic-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
	"github.com/wikylyu/mtop/tunnel"
	"github.com/wikylyu/mtop/tunnel/protocol/quic"
)

const AppName = "mtop"
const AppVersion = "0.0.2"

var AppCommit = ""

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppName + " is a network proxy server",

	Run: func(cmd *cobra.Command, args []string) {
		stype, listen, config := initServerConfig()
		if stype == "quic" {
			runQUICServer(listen, config)
		} else {
			runTLSServer(listen, config)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("show %s's version", AppName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version: %s\n", AppName, AppVersion)
		fmt.Printf("git commit: %s\n", AppCommit)
	},
}

var cfgFile string

func init() {
	cobra.OnInitialize(initMTop)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("config file (default is %s/etc/%s/%s.yaml)", config.PREFIX, AppName, AppName))
	rootCmd.AddCommand(versionCmd)
}

func initMTop() {
	config.Init(AppName, cfgFile)
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
		Proto  string `json:"proto" yaml:"proto"`
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

	if cfg.Proto == "" {
		cfg.Proto = "mtop"
	}

	return strings.ToLower(cfg.Type), cfg.Listen, &tls.Config{
		Certificates: []tls.Certificate{cer},
		NextProtos:   []string{cfg.Proto},
	}
}

func handleQUICConn(conn quicgo.Connection) {
	defer conn.CloseWithError(0, "")
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Debugf("[QUIC] accept stream error:%v", err)
			break
		}
		t := tunnel.NewTunnel(quic.NewConn(stream, conn.RemoteAddr()))
		go t.Run()
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
		go handleQUICConn(conn)
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
	rootCmd.Execute()
}
