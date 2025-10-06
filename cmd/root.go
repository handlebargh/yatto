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
	"fmt"
	"log"
	"os"

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

var configPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yatto",
	Short: "Interactive VCS-based todo-list for the command-line ",
	Run: func(_ *cobra.Command, _ []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("fatal error getting user home directory: %w", err))
		}

		config.InitConfig(home, &configPath)
		setCfg := config.Settings{
			ConfigPath: configPath,
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

		if viper.GetBool("git.remote.enable") || viper.GetBool("jj.remote.enable") {
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = s.Style.
				Foreground(lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}).
				Bold(true)

			fetchModel := fetchmodel.FetchModel{
				Spinner: s,
			}

			if _, err := tea.NewProgram(fetchModel, tea.WithAltScreen()).Run(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}
		}

		if _, err := tea.NewProgram(models.InitialProjectListModel(), tea.WithAltScreen()).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the config file")
}
