package main

import (
	"os"
	"os/signal"

	"github.com/wikylyu/mtop/climber/proxy"
	"github.com/wikylyu/mtop/config"
)

const (
	AppName = "climber"
)

func init() {
	config.Init(AppName)
	config.InitLog()
	proxy.Init()
}

func main() {
	go proxy.RunHTTPProxy()
	go proxy.RunSOCKS5Proxy()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
