package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/wikylyu/mtop/config"
	"github.com/wikylyu/mtop/db"
)

const (
	AppName = "mtop"
)

func init() {
	config.Init(AppName)
	// config.InitLog()
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
	var username, password, salt string
	flag.StringVar(&username, "username", "", "Username, can't be duplicated")
	flag.StringVar(&password, "password", "", "Plain password")
	flag.StringVar(&salt, "salt", "", "Password salt, leave it empty if you want to generate a random one automatically")
	flag.Parse()
	if username == "" || password == "" {
		flag.Usage()
		return
	}
	user, err := db.GetUserByUsername(username)
	if err != nil {
		panic(err)
	} else if user != nil {
		log.Infof("user %s already exists", username)
		return
	}
	user, err = db.CreateUser(username, password, salt)
	if err != nil {
		panic(err)
	}
	log.Infof("user %s created", user.Username)
}
