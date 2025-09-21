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
package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestCreateStorageDir(t *testing.T) {
	tests := []struct {
		name          string
		userInput     string
		expectCreated bool
		expectExit    bool
		expectPrompt  bool
	}{
		{
			name:          "user answers y",
			userInput:     "y\n",
			expectCreated: true,
		},
		{
			name:          "user answers Y",
			userInput:     "Y\n",
			expectCreated: true,
		},
		{
			name:          "user answers yes",
			userInput:     "yes\n",
			expectCreated: true,
		},
		{
			name:       "user answers n",
			userInput:  "n\n",
			expectExit: true,
		},
		{
			name:       "user answers random input",
			userInput:  "maybe\n",
			expectExit: true,
		},
		{
			name:       "user answers blank",
			userInput:  "\n",
			expectExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			storageDir := filepath.Join(tmp, "store")

			var out bytes.Buffer
			exited := false

			cfg := StorageDirConfig{
				Path:   storageDir,
				Stdin:  strings.NewReader(tt.userInput),
				Stdout: &out,
				Exit: func(int) {
					exited = true
				},
			}

			CreateStorageDir(cfg)

			// Check if directory was created
			_, err := os.Stat(storageDir)
			dirExists := !os.IsNotExist(err)

			if dirExists != tt.expectCreated {
				t.Errorf("expected directory created=%v, got %v", tt.expectCreated, dirExists)
			}

			if exited != tt.expectExit {
				t.Errorf("expected exit=%v, got %v", tt.expectExit, exited)
			}

			if !strings.Contains(out.String(), "Create storage directory") {
				t.Errorf("expected prompt in output, got %q", out.String())
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	storageDir := filepath.Join(tmp, "store")
	if err := os.MkdirAll(storageDir, 0o700); err != nil {
		t.Fatalf("failed to create storage dir: %v", err)
	}

	viper.Set("storage.path", storageDir)

	existingFile := filepath.Join(storageDir, "data.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		file     string
		expected bool
	}{
		{
			name:     "file exists",
			file:     "data.txt",
			expected: true,
		},
		{
			name:     "file does not exist",
			file:     "missing.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.file)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
