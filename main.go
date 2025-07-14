package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

type spinnerModel struct {
	spinner spinner.Model
	width   int
	height  int
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		git.PullCmd(),
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case git.GitPullDoneMsg:
		return m, tea.Quit

	case git.GitPullErrorMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m spinnerModel) View() string {
	content := fmt.Sprintf("%s Fetching data from remote…", m.spinner.View())

	// Center horizontally and vertically
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func initConfig(home string, configPath *string) {
	viper.SetDefault("storage.path", filepath.Join(home, ".yatto"))

	viper.SetDefault("git.default_branch", "main")
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

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = s.Style.
		Foreground(lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}).
		Bold(true)

	spinnerModel := spinnerModel{
		spinner: s,
	}

	if _, err := tea.NewProgram(spinnerModel, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(models.InitialTaskListModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
