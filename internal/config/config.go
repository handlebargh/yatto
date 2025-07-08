package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func CreateConfigFile() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(viper.GetString("config_dir"), 0755); err != nil {
				panic(fmt.Errorf("fatal error creating config directory: %w", err))
			}
			if err := viper.SafeWriteConfig(); err != nil {
				panic(fmt.Errorf("fatal error writing config file: %w", err))
			}
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}
