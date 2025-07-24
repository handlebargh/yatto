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

package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func CreateStorageDir() {
	storageDir := viper.GetString("storage.path")

	// Ask if storage directory should be created if it does not exist.
	_, err := os.Stat(storageDir)
	if os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Create storage directory at %s? [y|N]: ", storageDir)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occurred while reading input. Please try again", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "yes" || input == "y" {
			// Create storage directory.
			err := os.MkdirAll(storageDir, 0700)
			if err != nil {
				panic(fmt.Errorf("fatal error creating storage directory: %w", err))
			}
		} else {
			os.Exit(0)
		}
	}
}

func FileExists(file string) bool {
	fullPath := filepath.Join(viper.GetString("storage.path"), file)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}
