// Copyright 2025 handlebargh
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package main initializes and runs the Yatto TUI application.
// It handles configuration, git synchronization (optional), and loads the project list UI.
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

// red is an adaptive color used to indicate errors in the UI.
var red = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}

// spinnerModel defines the model used for displaying a spinner while syncing with a remote Git repository.
type spinnerModel struct {
	spinner spinner.Model
	err     error
	width   int
	height  int
}

// Init initializes the spinner model and starts the Git pull command.
func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		git.PullCmd(),
	)
}

// Update handles messages received during the spinner's lifecycle,
// such as window resize, spinner ticks, Git pull results, and user input.
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
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Interrupt
		}

		switch msg.String() {
		case "esc", "q":
			return m, tea.Interrupt
		}
	}

	return m, nil
}

// View renders the spinner UI, displaying a loading animation or an error message,
// centered in the terminal window.
func (m spinnerModel) View() string {
	var content string
	if m.err != nil {
		content = lipgloss.NewStyle().Foreground(red).Bold(true).Render("Error") +
			" fetching data from remote"
	} else {
		content = fmt.Sprintf("%s Fetching data from remote…", m.spinner.View())
	}

	// Center horizontally and vertically
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// initConfig sets default values for application configuration and
// attempts to load configuration from a file.
func initConfig(home string, configPath *string) {
	viper.SetDefault("storage.path", filepath.Join(home, ".yatto"))

	viper.SetDefault("git.default_branch", "main")
	viper.SetDefault("git.remote.enable", false)
	viper.SetDefault("git.remote.name", "origin")

	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(filepath.Join(home, ".config/yatto"))
	}
}

// main is the entry point of the Yatto application. It sets up configuration,
// creates storage directories, optionally performs a Git sync, and starts the TUI.
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
		s := spinner.New()
		s.Spinner = spinner.Dot
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
	}

	if _, err := tea.NewProgram(models.InitialProjectListModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
