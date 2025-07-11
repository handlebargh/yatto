package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func CreateConfigFile(home string) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(filepath.Join(home, ".config/yatto"), 0755); err != nil {
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
