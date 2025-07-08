package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	initConfig()
	config.CreateConfigFile()
	storage.CreateStorageDir()

	if viper.GetString("git_remote") != "" {
		err := git.GitPull(viper.GetString("storage_dir"))
		if err != nil {
			fmt.Println("Error syncing repository:", err)
			os.Exit(1)

		}
	}

	if _, err := tea.NewProgram(models.InitialListModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("fatal error getting user home directory: %w", err))
	}

	viper.SetDefault("config_dir", home+"/.config/yatto")
	viper.SetDefault("storage_dir", home+"/.local/share/yatto/tasks")
	viper.SetDefault("use_git", false)
	viper.SetDefault("git_remote", "")
	viper.SetDefault("push_on_change", false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config_dir"))
}
