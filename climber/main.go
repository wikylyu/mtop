package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/wikylyu/mtop/climber/proxy"
	"github.com/wikylyu/mtop/config"
)

const (
	AppName = "climber"
)

func initClimber(configFile string) {
	config.Init(AppName, configFile)
	config.InitLog()
	proxy.Init()
}

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.Parse()

	initClimber(configFile)

	go proxy.RunHTTPProxy()
	go proxy.RunSOCKS5Proxy()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
