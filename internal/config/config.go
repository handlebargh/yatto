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

	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/spf13/viper"
)

// ErrUserAborted is returned when a user cancels config file creation.
var ErrUserAborted = errors.New("user aborted config creation")

// Settings defines the runtime settings used by CreateConfigFile.
//
// Fields:
//   - Stdin:  Input stream used to read user responses (e.g., os.Stdin).
//   - Stdout: Output stream used to print prompts and messages (e.g., os.Stdout).
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
// with permissions 0755, and Viper writes the config file safely using
// viper.SafeWriteConfig.
//
// Returns ErrUserAborted if the user declines to create a config file, or an
// error if something goes wrong. A nil error means the config file was created
// successfully.
func CreateConfigFile(set Settings) error {
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("fatal error getting config: %w", err)
		}

		path := filepath.Join(set.Home, ".config", "config.toml")
		if set.ConfigPath != "" {
			path = set.ConfigPath
		}

		// Prompt for config file path
		_, err := helpers.PromptUser(
			set.Input,
			set.Output,
			fmt.Sprintf("Create config file as %s? [y|N]: ", path),
			"yes", "y", "Y",
		)
		if errors.Is(err, helpers.ErrUnexpectedInput) {
			return ErrUserAborted
		}
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Prompt for VCS
		inputVCS, err := helpers.PromptUser(
			set.Input,
			set.Output,
			"Which version control system would you like to use? [git(default)|jj]: ",
			"git", "jj", "",
		)
		if errors.Is(err, helpers.ErrUnexpectedInput) {
			return ErrUserAborted
		}
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if inputVCS == "" {
			inputVCS = "git"
		}
		viper.SetDefault("vcs.backend", inputVCS)

		// Create config dir
		if err := os.MkdirAll(filepath.Join(set.Home, ".config/yatto"), 0o755); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		// Write config file
		if err := viper.SafeWriteConfig(); err != nil {
			return fmt.Errorf("error writing config file: %w", err)
		}
	}

	return nil
}
