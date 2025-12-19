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

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

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
func InitConfig(v *viper.Viper, home string, configPath *string) {
	v.SetDefault("storage.path", filepath.Join(home, ".yatto"))

	// assignee
	v.SetDefault("assignee.show", false)
	v.SetDefault("assignee.show_printer", false)

	// author
	v.SetDefault("author.show", false)
	v.SetDefault("author.show_printer", false)

	// vcs
	v.SetDefault("vcs.backend", "git")

	// Git
	v.SetDefault("git.default_branch", "main")
	v.SetDefault("git.remote.enable", false)
	v.SetDefault("git.remote.name", "origin")

	// jj
	v.SetDefault("jj.default_branch", "main")
	v.SetDefault("jj.remote.enable", false)
	v.SetDefault("jj.remote.name", "origin")
	v.SetDefault("jj.remote.colocate", false)

	// colors
	v.SetDefault("colors.red_light", "#FE5F86")
	v.SetDefault("colors.red_dark", "#FE5F86")
	v.SetDefault("colors.vividred_light", "#FE134D")
	v.SetDefault("colors.vividred_dark", "#FE134D")
	v.SetDefault("colors.indigo_light", "#5A56E0")
	v.SetDefault("colors.indigo_dark", "#7571F9")
	v.SetDefault("colors.green_light", "#02BA84")
	v.SetDefault("colors.green_dark", "#02BF87")
	v.SetDefault("colors.orange_light", "#FFB733")
	v.SetDefault("colors.orange_dark", "#FFA336")
	v.SetDefault("colors.blue_light", "#1E90FF")
	v.SetDefault("colors.blue_dark", "#1E90FF")
	v.SetDefault("colors.yellow_light", "#CCCC00")
	v.SetDefault("colors.yellow_dark", "#CCCC00")
	v.SetDefault("colors.badge_text_light", "#000000")
	v.SetDefault("colors.badge_text_dark", "#000000")

	// Form themes
	v.SetDefault("colors.form.theme", "Base16")

	if *configPath != "" {
		v.SetConfigFile(*configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("toml")
		v.AddConfigPath(filepath.Join(home, ".config", "yatto"))
		*configPath = filepath.Join(home, ".config", "yatto", "config.toml")
	}
}

// Settings defines the runtime settings used by CreateConfigFile.
//
// Fields:
//   - Viper:  The viper instance to use for configuration.
//   - Input:  Input stream to read user responses (e.g., os.Stdin).
//   - Output: Output stream to print prompts and messages (e.g., os.Stdout).
type Settings struct {
	Viper      *viper.Viper
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
// The function then asks the user to choose some configuration values. These
// values are stored in the config file.
//
// If necessary, the configuration directory ($HOME/.config/yatto) is created
// with permissions 0750, and Viper writes the config file safely using
// viper.SafeWriteConfig.
//
// Returns an error if something goes wrong. A nil error means the config file
// was created successfully.
func CreateConfigFile(settings Settings) error {
	if err := settings.Viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("fatal error getting config: %w", err)
		}

		path := filepath.Join(settings.Home, ".config", "yatto", "config.toml")
		if settings.ConfigPath != "" {
			path = settings.ConfigPath
		}

		var (
			createConfig bool
			choiceVCS    string
			colocateJJ   bool
			remoteURL    string
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Create config file?").
					Description(fmt.Sprintf("Location: %s", path)).
					Affirmative("Yes").
					Negative("No").
					Value(&createConfig),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if !createConfig {
			return ErrUserAborted
		}

		form = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose your version control system").
					Options(
						huh.NewOption("Git", "git"),
						huh.NewOption("Jujutsu", "jj"),
					).
					Value(&choiceVCS),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		settings.Viper.Set("vcs.backend", choiceVCS)

		if choiceVCS == "jj" {
			form = huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Colocate JJ repository?").
						Affirmative("Yes").
						Negative("No").
						Value(&colocateJJ),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			settings.Viper.Set("jj.colocate", colocateJJ)
		}

		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Remote repository URL").
					Description("e.g. git@github.com:<username>/<repo>.git\nLeave empty to skip").
					Value(&remoteURL),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if remoteURL != "" {
			switch choiceVCS {
			case "git":
				settings.Viper.Set("git.remote.enable", true)
				settings.Viper.Set("git.remote.url", remoteURL)
			case "jj":
				settings.Viper.Set("jj.remote.enable", true)
				settings.Viper.Set("jj.remote.url", remoteURL)
			}
		}

		// Create config dir
		if err := os.MkdirAll(filepath.Join(settings.Home, ".config", "yatto"), 0o750); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		// Write config file
		if err := settings.Viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("error writing config file: %w", err)
		}
	}

	return nil
}

// LoadAndValidateConfig loads configuration values from viper and validates them.
// It returns an error if any configuration value is invalid or missing required fields.
// This function should be called at application startup after viper has been initialized.
func LoadAndValidateConfig(v *viper.Viper) error {
	cfg := &config{
		assigneeShow:        v.GetBool("assignee.show"),
		assigneeShowPrinter: v.GetBool("assignee.show_printer"),
		authorShow:          v.GetBool("author.show"),
		authorShowPrinter:   v.GetBool("author.show_printer"),
		gitRemoteEnable:     v.GetBool("git.remote.enable"),
		jjRemoteEnable:      v.GetBool("jj.remote.enable"),
		jjRemoteColocate:    v.GetBool("jj.remote.colocate"),
		storagePath:         v.GetString("storage.path"),
		vcsBackend:          v.GetString("vcs.backend"),
		gitDefaultBranch:    v.GetString("git.default_branch"),
		gitRemoteName:       v.GetString("git.remote.name"),
		jjDefaultBranch:     v.GetString("jj.default_branch"),
		jjRemoteName:        v.GetString("jj.remote.name"),
		colorsFormTheme:     v.GetString("colors.form.theme"),
		colorValues: map[string]string{
			"colors.red_light":        v.GetString("colors.red_light"),
			"colors.red_dark":         v.GetString("colors.red_dark"),
			"colors.vividred_light":   v.GetString("colors.vividred_light"),
			"colors.vividred_dark":    v.GetString("colors.vividred_dark"),
			"colors.indigo_light":     v.GetString("colors.indigo_light"),
			"colors.indigo_dark":      v.GetString("colors.indigo_dark"),
			"colors.green_light":      v.GetString("colors.green_light"),
			"colors.green_dark":       v.GetString("colors.green_dark"),
			"colors.orange_light":     v.GetString("colors.orange_light"),
			"colors.orange_dark":      v.GetString("colors.orange_dark"),
			"colors.blue_light":       v.GetString("colors.blue_light"),
			"colors.blue_dark":        v.GetString("colors.blue_dark"),
			"colors.yellow_light":     v.GetString("colors.yellow_light"),
			"colors.yellow_dark":      v.GetString("colors.yellow_dark"),
			"colors.badge_text_light": v.GetString("colors.badge_text_light"),
			"colors.badge_text_dark":  v.GetString("colors.badge_text_dark"),
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
