package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wikylyu/mtop/db"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var userAddParams struct {
	Username string
	Password string
	Salt     string
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Run: func(cmd *cobra.Command, args []string) {
		username := userAddParams.Username
		password := userAddParams.Password
		salt := userAddParams.Salt
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

var userDelParams struct {
	Username string
}

var userDelCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		username := userDelParams.Username
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
	userAddCmd.PersistentFlags().StringVarP(&userAddParams.Username, "username", "u", "", "Username")
	userAddCmd.PersistentFlags().StringVarP(&userAddParams.Password, "password", "p", "", "Password")
	userAddCmd.PersistentFlags().StringVarP(&userAddParams.Salt, "salt", "s", "", "Password salt used to encrypt")

	UserCmd.AddCommand(userAddCmd)

	userDelCmd.PersistentFlags().StringVarP(&userDelParams.Username, "username", "u", "", "Username")
	UserCmd.AddCommand(userDelCmd)
}
