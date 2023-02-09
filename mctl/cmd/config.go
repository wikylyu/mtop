package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/tunnel/protocol/mtop"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Config management",

	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here

	},
}

var configURLParam struct {
	Username string
	Password string
	Host     string
	Port     uint16
	Type     string
	Proto    string
}

var configURLCmd = &cobra.Command{
	Use:   "url",
	Short: "Generate url for special user config",

	Run: func(cmd *cobra.Command, args []string) {
		username := configURLParam.Username
		password := configURLParam.Password
		host := configURLParam.Host
		port := configURLParam.Port
		stype := configURLParam.Type
		proto := configURLParam.Proto
		u := mtop.GenerateMTopURL(username, password, host, port, stype, proto)
		fmt.Printf("%s\n", u.String())
	},
}

func init() {
	configURLCmd.PersistentFlags().StringVarP(&configURLParam.Username, "username", "u", "", "Username")
	configURLCmd.PersistentFlags().StringVarP(&configURLParam.Password, "password", "p", "", "Password")
	configURLCmd.PersistentFlags().StringVarP(&configURLParam.Host, "host", "H", "", "Server hostname")
	configURLCmd.PersistentFlags().Uint16VarP(&configURLParam.Port, "port", "P", 0, "Server port")
	configURLCmd.PersistentFlags().StringVarP(&configURLParam.Type, "type", "t", "tls", "Transport type")
	configURLCmd.PersistentFlags().StringVarP(&configURLParam.Proto, "proto", "o", "mtop", "Proto name, strongly recommended to set a custom one")
	ConfigCmd.AddCommand(configURLCmd)
}
