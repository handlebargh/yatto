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

package cmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/fetchmodel"
	"github.com/handlebargh/yatto/internal/printer"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pullFlag      bool
	authorFlag    bool
	assigneeFlag  bool
	printProjects string
	printRegex    string
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print tasks to stdout",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		_, gitErr := exec.LookPath("git")
		_, jjErr := exec.LookPath("jj")
		if gitErr != nil && jjErr != nil {
			return errors.New("yatto requires either 'git' or 'jj' to be installed")
		}

		return nil
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		setCfg := config.Settings{
			Viper:      appConfig.Viper,
			ConfigPath: configPath,
			Home:       homePath,
			Input:      os.Stdin,
			Output:     os.Stdout,
			Exit:       os.Exit,
		}

		if err := config.CreateConfigFile(setCfg); err != nil {
			if errors.Is(err, config.ErrUserAborted) {
				os.Exit(0)
			}
			return err
		}

		err := config.LoadAndValidateConfig(setCfg.Viper)
		if err != nil {
			return err
		}

		setStorage := storage.Settings{
			Viper:  appConfig.Viper,
			Input:  os.Stdin,
			Output: os.Stdout,
			Exit:   os.Exit,
		}

		if err := storage.CreateStorageDir(setStorage); err != nil {
			if errors.Is(err, storage.ErrUserAborted) {
				os.Exit(0)
			}
			return err
		}

		if pullFlag &&
			((appConfig.Viper.GetString("vcs.backend") == "git" && appConfig.Viper.GetBool("git.remote.enable")) ||
				(appConfig.Viper.GetString("vcs.backend") == "jj" && appConfig.Viper.GetBool("jj.remote.enable"))) {
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = s.Style.
				Foreground(lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}).
				Bold(true)

			fetchModel := fetchmodel.FetchModel{
				Spinner: s,
			}

			if _, err := tea.NewProgram(fetchModel, tea.WithAltScreen()).Run(); err != nil {
				return err
			}
		}

		printTaskList(appConfig.Viper, printProjects, printRegex)

		return nil
	},
}

// printTaskList prints a list of tasks based on the provided project names
// and a regular expression filter.
//
// The function takes three arguments as input:
// - v: a viper instance for configuration.
// - printProjects: a space-separated string of project names.
// - printRegex: a regular expression used to filter tasks.
//
// It splits the printProjects string into individual project names, then calls
// printer.PrintTasks with the regex and the list of projects.
func printTaskList(v *viper.Viper, printProjects, printRegex string) {
	// Get a slice of strings from user input.
	projects := strings.Fields(printProjects)

	printer.PrintTasks(v, printRegex, authorFlag, assigneeFlag, projects...)
}

func init() {
	printCmd.Flags().BoolVarP(&pullFlag, "pull", "p", false, "Pull the remote before printing")
	printCmd.Flags().BoolVarP(&authorFlag, "author", "a", false, "Print tasks only authored by you")
	printCmd.Flags().BoolVarP(&assigneeFlag, "assignee", "A", false, "Print tasks only assigned to you")
	printCmd.Flags().StringVarP(&printProjects, "projects", "P", "", "List of project UUIDs to print from")
	printCmd.Flags().StringVarP(&printRegex, "regex", "r", "", "Regex to filter task labels")
	rootCmd.AddCommand(printCmd)
}
