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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config defines the runtime dependencies and settings used by
// functions that manage the storage directory.
//
// Fields:
//   - Path:   Filesystem path to the storage directory.
//   - Stdin:  Input stream used to read user responses (e.g., os.Stdin).
//   - Stdout: Output stream used to print prompts and messages (e.g., os.Stdout).
//   - Exit:   Function invoked to terminate the process (e.g., os.Exit).
type Config struct {
	Path   string
	Stdin  io.Reader
	Stdout io.Writer
	Exit   func(int)
}

// CreateStorageDir checks if the configured storage directory exists,
// and prompts the user to create it if it does not. If the user confirms,
// the directory is created with 0700 permissions. Exits the program if the
// user declines or an error occurs during input.
func CreateStorageDir(cfg Config) {
	storageDir := cfg.Path

	_, err := os.Stat(storageDir)
	if os.IsNotExist(err) {
		reader := bufio.NewReader(cfg.Stdin)

		_, err := fmt.Fprintf(cfg.Stdout, "Create storage directory at %s? [y|N]: ", storageDir)
		if err != nil {
			return
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			_, err := fmt.Fprintln(cfg.Stdout, "An error occurred while reading input. Please try again", err)
			if err != nil {
				return
			}

			return
		}

		input = strings.TrimSpace(input)

		if input == "yes" || input == "y" || input == "Y" {
			if err := os.MkdirAll(storageDir, 0o700); err != nil {
				panic(fmt.Errorf("fatal error creating storage directory: %w", err))
			}
		} else {
			cfg.Exit(0)
		}
	}
}

// FileExists returns true if the specified file exists within the configured
// storage directory. It uses os.Stat to check for existence and ignores other errors.
func FileExists(file string) bool {
	fullPath := filepath.Join(viper.GetString("storage.path"), file)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}
