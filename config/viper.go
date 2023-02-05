package config

import (
	"path"

	"github.com/spf13/viper"
)

var PREFIX string

func ViperInit(appname, configFile string) error {
	if configFile == "" {
		viper.SetConfigName(appname)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		if PREFIX != "" {
			viper.AddConfigPath(path.Join(PREFIX, "etc", appname))
		}
	} else {
		viper.SetConfigFile(configFile)
	}
	return viper.ReadInConfig() // Find and read the config file
}

func ViperUnmarshal(k string, v interface{}) error {
	return viper.UnmarshalKey(k, v)
}
