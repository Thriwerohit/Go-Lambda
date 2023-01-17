package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config/")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("file not found")
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
		}
	}

	return v
}
