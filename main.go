package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("fatal error getting user home directory: %w", err))
	}

	initConfig(home, configPath)
	config.CreateConfigFile(home)
	storage.CreateStorageDir()

	if viper.GetBool("git.remote.enable") {
		fmt.Println("Synchronizing tasks...")
		err := git.PullAll()
		if err != nil {
			fmt.Println("Error syncing repository:", err)
			os.Exit(1)
		}
	}

	if _, err := tea.NewProgram(models.InitialTaskListModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func initConfig(home string, configPath *string) {
	viper.SetDefault("storage.path", filepath.Join(home, ".yatto"))

	viper.SetDefault("git.enable", true)
	viper.SetDefault("git.remote.enable", false)
	viper.SetDefault("git.remote.name", "origin")
	viper.SetDefault("git.remote.push_on_commit", false)

	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(filepath.Join(home, ".config/yatto"))
	}
}
