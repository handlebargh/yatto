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

// Package config provides the logic to create the configuration file.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/spf13/viper"
)

// foregroundColor is used in the init dialog to highlight options for instance.
const foregroundColor = lipgloss.Color("#c71585")

var (
	// ErrUserAborted is returned when a user cancels config file creation.
	ErrUserAborted = errors.New("user aborted config creation")

	// branchNameRegexp validates branch names to prevent command injection.
	branchNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9./_-]+$`)

	// remoteNameRegexp validates remote names to prevent command injection.
	remoteNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Validates color codes
	colorRegexp = regexp.MustCompile(`^#?[a-fA-F0-9]+$`)
)

// config is used to load all values from the configuration file
// in order to validate them.
type config struct {
	assigneeShow        bool
	assigneeShowPrinter bool
	authorShow          bool
	authorShowPrinter   bool
	gitRemoteEnable     bool
	jjRemoteEnable      bool
	jjRemoteColocate    bool
	storagePath         string
	vcsBackend          string
	gitDefaultBranch    string
	gitRemoteName       string
	jjDefaultBranch     string
	jjRemoteName        string
	colorsFormTheme     string
	colorValues         map[string]string
}

// InitConfig sets default values for application configuration and
// attempts to load configuration from a file.
func InitConfig(home string, configPath *string) {
	viper.SetDefault("storage.path", filepath.Join(home, ".yatto"))

	// assignee
	viper.SetDefault("assignee.show", false)
	viper.SetDefault("assignee.show_printer", false)

	// author
	viper.SetDefault("author.show", false)
	viper.SetDefault("author.show_printer", false)

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
	viper.SetDefault("jj.remote.colocate", false)

	// colors
	viper.SetDefault("colors.red_light", "#FE5F86")
	viper.SetDefault("colors.red_dark", "#FE5F86")
	viper.SetDefault("colors.vividred_light", "#FE134D")
	viper.SetDefault("colors.vividred_dark", "#FE134D")
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
		viper.AddConfigPath(filepath.Join(home, ".config", "yatto"))
		*configPath = filepath.Join(home, ".config", "yatto", "config.toml")
	}
}

// Settings defines the runtime settings used by CreateConfigFile.
//
// Fields:
//   - Input:  Input stream used to read user responses (e.g., os.Stdin).
//   - Output: Output stream used to print prompts and messages (e.g., os.Stdout).
type Settings struct {
	ConfigPath string
	Home       string
	Input      io.Reader
	Output     io.Writer
	Exit       func(int)
}

// CreateConfigFile ensures that a configuration file exists for the application.
// It first attempts to read an existing config using Viper. If no config file
// is found, the user is prompted to confirm creation of a new one at the default
// location ($HOME/.config/config.toml) or at set.ConfigPath if specified.
//
// The function then asks the user to choose a version control system (VCS),
// defaulting to "git" if none is provided. The selected VCS backend is stored
// as a default value in the config.
//
// If necessary, the configuration directory ($HOME/.config/yatto) is created
// with permissions 0750, and Viper writes the config file safely using
// viper.SafeWriteConfig.
//
// Returns an error if something goes wrong. A nil error means the config file
// was created successfully.
func CreateConfigFile(set Settings) error {
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("fatal error getting config: %w", err)
		}

		path := filepath.Join(set.Home, ".config", "yatto", "config.toml")
		if set.ConfigPath != "" {
			path = set.ConfigPath
		}

		hexagon := lipgloss.NewStyle().
			Foreground(foregroundColor).
			Render("⬢")

		// Prompt for config file path
		yesOrNo, err := helpers.PromptUser(
			set.Input,
			set.Output,
			fmt.Sprintf("\n%s Create config file at %s ? %s: ",
				hexagon,
				lipgloss.NewStyle().Bold(true).Render(path),
				lipgloss.NewStyle().Bold(true).Foreground(foregroundColor).Render("[y|N]"),
			),
			"yes", "y", "no", "n",
		)
		if yesOrNo == "no" || yesOrNo == "n" {
			return ErrUserAborted
		}
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Prompt for VCS
		inputVCS, err := helpers.PromptUser(
			set.Input,
			set.Output,
			fmt.Sprintf(
				"\n%s Which %s would you like to use?\n\n  1) %s (default)\n  2) %s\n\nEnter choice %s: ",
				hexagon,
				lipgloss.NewStyle().Bold(true).Foreground(colors.Green()).Render("version control system"),
				lipgloss.NewStyle().Bold(true).Foreground(colors.Green()).Render("Git"),
				lipgloss.NewStyle().Bold(true).Foreground(colors.Green()).Render("Jujutsu"),
				lipgloss.NewStyle().Bold(true).Foreground(foregroundColor).Render("[1|2]"),
			),
			"1",
			"2",
			"",
		)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if inputVCS == "1" || inputVCS == "" {
			inputVCS = "git"
		}

		if inputVCS == "2" {
			inputVCS = "jj"

			// Prompt for colocation
			yesOrNo, err := helpers.PromptUser(
				set.Input,
				set.Output,
				fmt.Sprintf("\n%s Do you want to %s the jj repository? %s: ",
					hexagon,
					lipgloss.NewStyle().Bold(true).Foreground(colors.Green()).Render("colocate"),
					lipgloss.NewStyle().Bold(true).Foreground(foregroundColor).Render("[y|N]"),
				),
				"yes", "y", "no", "n",
			)
			if yesOrNo == "no" || yesOrNo == "n" {
				return ErrUserAborted
			}

			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}

			viper.Set("jj.colocate", true)
		}

		viper.Set("vcs.backend", inputVCS)

		// Prompt for remote URL
		inputRemote, err := helpers.PromptUser(
			set.Input,
			set.Output,
			fmt.Sprintf(
				"\n%s Enter your remote repository URL (e.g. git@github.com:<username>/<repo>.git | leave %s): ",
				hexagon,
				lipgloss.NewStyle().Bold(true).Foreground(foregroundColor).Render("empty to skip"),
			),
		)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if inputRemote != "" && inputVCS == "git" {
			viper.Set("git.remote.enable", true)
			viper.Set("git.remote.url", inputRemote)
		} else if inputRemote != "" && inputVCS == "jj" {
			viper.Set("jj.remote.enable", true)
			viper.Set("jj.remote.url", inputRemote)
		}

		// Create config dir
		if err := os.MkdirAll(filepath.Join(set.Home, ".config", "yatto"), 0o750); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		// Write config file
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("error writing config file: %w", err)
		}
	}

	return nil
}

// LoadAndValidateConfig loads configuration values from viper and validates them.
// It returns an error if any configuration value is invalid or missing required fields.
// This function should be called at application startup after viper has been initialized.
func LoadAndValidateConfig() error {
	cfg := &config{
		assigneeShow:        viper.GetBool("assignee.show"),
		assigneeShowPrinter: viper.GetBool("assignee.show_printer"),
		authorShow:          viper.GetBool("author.show"),
		authorShowPrinter:   viper.GetBool("author.show_printer"),
		gitRemoteEnable:     viper.GetBool("git.remote.enable"),
		jjRemoteEnable:      viper.GetBool("jj.remote.enable"),
		jjRemoteColocate:    viper.GetBool("jj.remote.colocate"),
		storagePath:         viper.GetString("storage.path"),
		vcsBackend:          viper.GetString("vcs.backend"),
		gitDefaultBranch:    viper.GetString("git.default_branch"),
		gitRemoteName:       viper.GetString("git.remote.name"),
		jjDefaultBranch:     viper.GetString("jj.default_branch"),
		jjRemoteName:        viper.GetString("jj.remote.name"),
		colorsFormTheme:     viper.GetString("colors.form.theme"),
		colorValues: map[string]string{
			"colors.red_light":        viper.GetString("colors.red_light"),
			"colors.red_dark":         viper.GetString("colors.red_dark"),
			"colors.vividred_light":   viper.GetString("colors.vividred_light"),
			"colors.vividred_dark":    viper.GetString("colors.vividred_dark"),
			"colors.indigo_light":     viper.GetString("colors.indigo_light"),
			"colors.indigo_dark":      viper.GetString("colors.indigo_dark"),
			"colors.green_light":      viper.GetString("colors.green_light"),
			"colors.green_dark":       viper.GetString("colors.green_dark"),
			"colors.orange_light":     viper.GetString("colors.orange_light"),
			"colors.orange_dark":      viper.GetString("colors.orange_dark"),
			"colors.blue_light":       viper.GetString("colors.blue_light"),
			"colors.blue_dark":        viper.GetString("colors.blue_dark"),
			"colors.yellow_light":     viper.GetString("colors.yellow_light"),
			"colors.yellow_dark":      viper.GetString("colors.yellow_dark"),
			"colors.badge_text_light": viper.GetString("colors.badge_text_light"),
			"colors.badge_text_dark":  viper.GetString("colors.badge_text_dark"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate checks that all configuration values are valid and consistent.
// It validates storage paths, VCS backend settings (git/jj), branch and remote names
// to prevent command injection, form theme names, and color codes.
// Returns an error describing the first validation failure encountered.
func (c *config) Validate() error {
	// Storage path validation
	if c.storagePath == "" {
		return fmt.Errorf("storage path cannot be empty")
	}
	if !filepath.IsAbs(c.storagePath) {
		return fmt.Errorf("storage path must be absolute: %q", c.storagePath)
	}

	// VCS backend validation
	switch c.vcsBackend {
	case "git":
		if !branchNameRegexp.MatchString(c.gitDefaultBranch) {
			return fmt.Errorf("invalid branch name: %q", c.gitDefaultBranch)
		}
		if !remoteNameRegexp.MatchString(c.gitRemoteName) {
			return fmt.Errorf("invalid remote name: %q", c.gitRemoteName)
		}
	case "jj":
		if !branchNameRegexp.MatchString(c.jjDefaultBranch) {
			return fmt.Errorf("invalid branch name: %q", c.jjDefaultBranch)
		}
		if !remoteNameRegexp.MatchString(c.jjRemoteName) {
			return fmt.Errorf("invalid remote name: %q", c.jjRemoteName)
		}
	default:
		return fmt.Errorf("unknown vcs backend: %s", c.vcsBackend)
	}

	// Form theme validation
	validThemes := map[string]bool{
		"Charm":      true,
		"Dracula":    true,
		"Catppuccin": true,
		"Base16":     true,
		"Base":       true,
	}

	if !validThemes[c.colorsFormTheme] {
		return fmt.Errorf(
			"unknown colors.form.theme: %s (valid: Charm, Dracula, Catppuccin, Base16, Base)",
			c.colorsFormTheme,
		)
	}

	// Color values validation
	for k, v := range c.colorValues {
		if !colorRegexp.MatchString(v) {
			return fmt.Errorf("invalid color value for '%s': %q", k, v)
		}
	}

	return nil
}
