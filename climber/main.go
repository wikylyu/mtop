package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/climber/proxy"
	"github.com/wikylyu/mtop/config"
)

const AppName = "climber"
const AppVersion = "0.0.2"

var AppCommit = ""

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppName + " is a client agent for mtop server",

	Run: func(cmd *cobra.Command, args []string) {
		go proxy.RunHTTPProxy()
		go proxy.RunSOCKS5Proxy()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
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
	cobra.OnInitialize(initClimber)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("config file (default is %s/etc/%s/%s.yaml)", config.PREFIX, AppName, AppName))
	rootCmd.AddCommand(versionCmd)
}

func initClimber() {
	config.Init(AppName, cfgFile)
	config.InitLog()
	proxy.Init()
}

func main() {
	go func() {
		http.ListenAndServe("127.0.0.1:1616", nil)
	}()
	rootCmd.Execute()
}
