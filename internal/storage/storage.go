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

// Package storage provides the logic to create the storage directory.
package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

// ErrUserAborted is returned when a user cancels storage directory creation.
var ErrUserAborted = errors.New("user aborted config creation")

// Settings defines settings used by CreateStorageDir.
//
// Fields:
//   - Viper:  The viper instance to use for configuration.
//   - Path:   Filesystem path to the storage directory.
//   - Input:  Input stream used to read user responses (e.g., os.Stdin).
//   - Output: Output stream used to print prompts and messages (e.g., os.Stdout).
//   - Exit:   Function invoked to terminate the process (e.g., os.Exit).
type Settings struct {
	Viper  *viper.Viper
	Input  io.Reader
	Output io.Writer
	Exit   func(int)
}

// CreateStorageDir checks if the configured storage directory exists,
// and prompts the user to create it if it does not. If the user confirms,
// the directory is created with 0700 permissions. Exits the program if the
// user declines or an error occurs during input.
func CreateStorageDir(settings Settings) error {
	storageDir := settings.Viper.GetString("storage.path")

	_, err := os.Stat(storageDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		var createStorage bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Create storage directory?").
					Description(fmt.Sprintf("Location: %s", storageDir)).
					Affirmative("Yes").
					Negative("No").
					Value(&createStorage),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if !createStorage {
			return ErrUserAborted
		}

		backend := settings.Viper.GetString("vcs.backend")
		if settings.Viper.GetBool(backend + ".remote.enable") {

			jjCmd := []string{
				"jj",
				"git",
				"clone",
				"--remote",
				settings.Viper.GetString("jj.remote.name"),
				settings.Viper.GetString("jj.remote.url"),
				storageDir,
			}

			if settings.Viper.GetBool("jj.remote.colocate") {
				jjCmd = append(jjCmd, "--colocate")
			}

			cmds := map[string][]string{
				"git": {"git", "clone", settings.Viper.GetString("git.remote.url"), storageDir},
				"jj":  jjCmd,
			}

			args, ok := cmds[backend]
			if ok {
				cmd := exec.Command(args[0], args[1:]...) // #nosec G204 Command uses validated config values

				cmd.Stdout = settings.Output
				cmd.Stderr = settings.Output

				if err := cmd.Run(); err != nil {
					return err
				}
			}

			// Rename branch if it's not our default.
			if backend == "git" {
				moveCmd := exec.Command("git", // #nosec G204 Command uses validated config value
					"branch",
					"--move", settings.Viper.GetString("git.default_branch"),
				)
				moveCmd.Dir = storageDir
				if err := moveCmd.Run(); err != nil {
					return err
				}
			}
		} else {
			if err := os.MkdirAll(storageDir, 0o700); err != nil {
				return fmt.Errorf("fatal error creating storage directory: %w", err)
			}
		}
	}

	return nil
}

// FileExists returns true if the specified file exists within the configured
// storage directory. It uses os.Stat to check for existence and ignores other errors.
func FileExists(v *viper.Viper, file string) bool {
	fullPath := path.Join(v.GetString("storage.path"), file)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}
