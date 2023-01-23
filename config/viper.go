package config

import (
	"github.com/spf13/viper"
)

func ViperInit(appname string) error {
	viper.SetConfigName(appname)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	return viper.ReadInConfig() // Find and read the config file
}

func ViperUnmarshal(k string, v interface{}) error {
	return viper.UnmarshalKey(k, v)
}
