// Copyright 2025 handlebargh and contributors
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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/handlebargh/yatto/internal/printer"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
)

var (
	// red is an adaptive color used to indicate errors in the UI.
	red = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}

	// revision saves the git commit the application is built from.
	revision = "unknown"

	// revisionDate saves the commit's date.
	revisionDate = "unknown"

	// goVersion saves the Go version the application is built with.
	goVersion = runtime.Version()
)

// versionInfo returns the version, commit sha hash and commit date
// from which the application is built.
func versionInfo() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "Unable to read version information."
	}

	if buildInfo.Main.Version != "" {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.time":
				revisionDate = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					revision += "+dirty"
				}
			}
		}

		return fmt.Sprintf("Version:\t%s\nRevision:\t%s\nRevisionDate:\t%s\nGoVersion:\t%s\n",
			buildInfo.Main.Version, revision, revisionDate, goVersion)
	}

	return fmt.Sprintf(
		"Version:\tunknown\nRevision:\tunknown\nRevisionDate:\tunknown\nGoVersion:\t%s\n",
		goVersion,
	)
}

// versionHeader returns the stylized application name
// and project URL.
func versionHeader() string {
	return `
 ____ ____ ____ ____ ____ 
||y |||a |||t |||t |||o ||
||__|||__|||__|||__|||__||
|/__\|/__\|/__\|/__\|/__\|

https://github.com/handlebargh/yatto
`
}

// printTaskList prints a list of tasks based on the provided project names
// and a regular expression filter.
//
// The function takes two strings as input:
// - printProjects: a space-separated string of project names.
// - printRegex: a regular expression used to filter tasks.
//
// It splits the printProjects string into individual project names, then calls
// printer.PrintTasks with the regex and the list of projects.
func printTaskList(printProjects, printRegex string) {
	// Get a slice of strings from user input.
	projects := strings.Fields(printProjects)

	printer.PrintTasks(printRegex, projects...)
}

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
		vcs.PullCmd(),
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

	case vcs.PullDoneMsg:
		return m, tea.Quit

	case vcs.PullErrorMsg:
		m.err = msg.Err
		return m, nil

	case vcs.PullNoInitMsg:
		m.err = vcs.ErrorNoInit
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
		if errors.Is(m.err, vcs.ErrorNoInit) {
			content = lipgloss.NewStyle().Foreground(red).Bold(true).Render("Error ") +
				m.err.Error()
		} else {
			content = lipgloss.NewStyle().Foreground(red).Bold(true).Render("Error") +
				" fetching data from remote"
		}
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

	// vcs
	viper.SetDefault("vcs.backend", "git")

	// Git
	viper.SetDefault("git.default_branch", "main")
	viper.SetDefault("git.remote.enable", false)
	viper.SetDefault("git.remote.name", "origin")

	// jj
	viper.SetDefault("jj.default_branch", "main")
	viper.SetDefault("jj.remote.enable", false)
	viper.SetDefault("jj.remote.name", "origin")

	// colors
	viper.SetDefault("colors.red_light", "#FE5F86")
	viper.SetDefault("colors.red_dark", "#FE5F86")
	viper.SetDefault("colors.vividRed_light", "#FE134D")
	viper.SetDefault("colors.vividRed_dark", "#FE134D")
	viper.SetDefault("colors.indigo_light", "#5A56E0")
	viper.SetDefault("colors.indigo_dark", "#7571F9")
	viper.SetDefault("colors.green_light", "#02BA84")
	viper.SetDefault("colors.green_dark", "#02BF87")
	viper.SetDefault("colors.orange_light", "#FFB733")
	viper.SetDefault("colors.orange_dark", "#FFA336")
	viper.SetDefault("colors.blue_light", "#1E90FF")
	viper.SetDefault("colors.blue_dark", "#1E90FF")
	viper.SetDefault("colors.yellow_light", "#CCCC00")
	viper.SetDefault("colors.yellow_dark", "#CCCC00")
	viper.SetDefault("colors.badge_text_light", "#000000")
	viper.SetDefault("colors.badge_text_dark", "#000000")

	// Form themes
	viper.SetDefault("colors.form.theme", "Base16")

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
	versionFlag := flag.Bool("version", false, "Print application version")
	printFlag := flag.Bool("print", false, "Print tasks to stdout")
	pullFlag := flag.Bool("pull", false, "Pull the remote before printing")
	printProjects := flag.String("projects", "", "List of project UUIDs to print from")
	printRegex := flag.String("regex", "", "Regex to be used on task labels")
	flag.Parse()

	if *versionFlag {
		fmt.Println(versionHeader())
		fmt.Println(versionInfo())
		os.Exit(0)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("fatal error getting user home directory: %w", err))
	}

	initConfig(home, configPath)
	setCfg := config.Settings{
		ConfigPath: *configPath,
		Home:       home,
		Input:      os.Stdin,
		Output:     os.Stdout,
		Exit:       os.Exit,
	}

	if err := config.CreateConfigFile(setCfg); err != nil {
		if errors.Is(err, config.ErrUserAborted) {
			os.Exit(0)
		}
		log.Fatalf("failed to create config: %v", err)
	}

	// Enforce valid vcs backend
	switch viper.GetString("vcs.backend") {
	case "git":
		break
	case "jj":
		break
	default:
		panic(fmt.Errorf("unknown vcs backend: %s", viper.GetString("vcs.backend")))
	}

	setStorage := storage.Settings{
		Path:   viper.GetString("storage.path"),
		Input:  os.Stdin,
		Output: os.Stdout,
		Exit:   os.Exit,
	}

	if err := storage.CreateStorageDir(setStorage); err != nil {
		if errors.Is(err, storage.ErrUserAborted) {
			os.Exit(0)
		}
		log.Fatalf("failed to create storage directory: %v", err)
	}

	// Print task list without pulling first.
	if *printFlag && !*pullFlag {
		printTaskList(*printProjects, *printRegex)
		os.Exit(0)
	}

	if viper.GetBool("git.remote.enable") || viper.GetBool("jj.remote.enable") {
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

	// Print task list after pulling.
	if *printFlag && *pullFlag {
		printTaskList(*printProjects, *printRegex)
		os.Exit(0)
	}

	if _, err := tea.NewProgram(models.InitialProjectListModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
