package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
	"github.com/wikylyu/mtop/mctl/cmd"
)

const AppName = "mctl"
const AppVersion = "0.0.2"

var AppCommit = ""

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: AppName + " is a command used to config mtop",

	Run: func(cmd *cobra.Command, args []string) {
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("config file (default is %s/etc/mtop/mtop.yaml)", config.PREFIX))
	rootCmd.AddCommand(versionCmd, cmd.UserCmd)
}

func initConfig() {
	config.Init("mtop", cfgFile)
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
	rootCmd.Execute()
}
