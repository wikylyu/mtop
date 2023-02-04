package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/db"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",

	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
var username, password, salt string

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			cmd.Usage()
			return
		}
		user, err := db.GetUserByUsername(username)
		if err != nil {
			return
		} else if user != nil {
			fmt.Printf("user %s already exists\n", user.Username)
			return
		}
		user, err = db.CreateUser(username, password, salt)
		if err != nil {
			return
		}
		fmt.Printf("user %s created\n", user.Username)
	},
}

var userDelCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			cmd.Usage()
			return
		}
		user, err := db.GetUserByUsername(username)
		if err != nil {
			return
		} else if user == nil {
			fmt.Printf("user %s not exists\n", username)
			return
		}
		if err := db.DeleteUser(username); err != nil {
			return
		}
		fmt.Printf("user %s deleted\n", username)
	},
}

func init() {
	userAddCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username")
	userAddCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password")
	userAddCmd.PersistentFlags().StringVarP(&salt, "salt", "s", "", "Password salt used to encrypt")

	userCmd.AddCommand(userAddCmd)

	userDelCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username")
	userCmd.AddCommand(userDelCmd)
}
