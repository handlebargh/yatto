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

// Package cmd provides the cobra commands and sub-commands.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/fetchmodel"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configPath string
	homePath   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yatto",
	Short: "Interactive VCS-based todo-list for the command-line",
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

		err := config.LoadAndValidateConfig()
		if err != nil {
			return err
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
			return err
		}

		if (viper.GetString("vcs.backend") == "git" && viper.GetBool("git.remote.enable")) ||
			(viper.GetString("vcs.backend") == "jj" && viper.GetBool("jj.remote.enable")) {
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

		if _, err := tea.NewProgram(models.InitialProjectListModel(), tea.WithAltScreen()).Run(); err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	var homeErr error
	homePath, homeErr = os.UserHomeDir()
	if homeErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fatal error getting user home directory: %v\n", homeErr)
		os.Exit(1)
	}

	config.InitConfig(homePath, &configPath)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the config file")
}
