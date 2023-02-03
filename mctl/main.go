package main

import (
	"flag"
	"fmt"
	"os"

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

	if len(os.Args) < 2 {
		help(os.Args[0])
		return
	}
	cmd := os.Args[1]
	switch cmd {
	case "user-add":
		userAdd(os.Args[2:])
	case "user-del":
		userDel(os.Args[2:])
	default:
		help(os.Args[0])
	}

}

func userAdd(args []string) {
	var username, password, salt string
	var update bool
	f := flag.NewFlagSet("user-add", flag.ExitOnError)
	f.StringVar(&username, "username", "", "Username, can't be duplicated")
	f.StringVar(&password, "password", "", "Plain password")
	f.BoolVar(&update, "update", false, "update password.")
	f.StringVar(&salt, "salt", "", "Password salt, leave it empty if you want to generate a random one automatically")
	f.Parse(args)
	if username == "" || password == "" {
		f.Usage()
		return
	}
	user, err := db.GetUserByUsername(username)
	if err != nil {
		panic(err)
	} else if user != nil && !update {
		log.Infof("user %s already exists", username)
		return
	}
	user, err = db.CreateUser(username, password, salt, update)
	if err != nil {
		panic(err)
	}
	log.Infof("user %s created", user.Username)
}

func userDel(args []string) {
	var username string
	f := flag.NewFlagSet("user-add", flag.ExitOnError)
	f.StringVar(&username, "username", "", "Username")
	f.Parse(args)
	if username == "" {
		f.Usage()
		return
	}
	if err := db.DeleteUser(username); err != nil {
		panic(err)
	}
	log.Infof("user %s deleted", username)
}

func help(appname string) {
	fmt.Printf("%s [command]\n\n", appname)
	fmt.Printf("available commands:\n")
	fmt.Printf("\tuser-add\tAdd a user\n")
	fmt.Printf("\tuser-del\tDelete a user\n")
	fmt.Printf("\n")
}
