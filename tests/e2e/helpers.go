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

// Package e2e contains all end to end tests.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// setGitAppConfig initializes a fresh git repo for testing and sets the viper
// config accordingly.
// Return the path to the testing storage directory.
func setGitAppConfig(t *testing.T) *viper.Viper {
	t.Helper()
	storagePath := setupGitRepo(t)
	v := viper.New()

	v.Set("storage.path", storagePath)
	v.Set("vcs.backend", "git")

	return v
}

// setJJAppConfig initializes a fresh jj repo for testing and sets the viper
// config accordingly.
// Return the path to the testing storage directory.
func setJJAppConfig(t *testing.T) *viper.Viper {
	t.Helper()
	storagePath := setupJJRepo(t)
	v := viper.New()

	v.Set("storage.path", storagePath)
	v.Set("vcs.backend", "jj")

	return v
}

// setupGitRepo creates a temporary directory and initializes a fresh git repo.
// It returns the path to the repo and ensures local git configs don't interfere.
func setupGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	runCmd(t, tmpDir, "git", "init", "--initial-branch", "main")
	runCmd(t, tmpDir, "git", "config", "user.name", "Test User")
	runCmd(t, tmpDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, tmpDir, "git", "config", "commit.gpgSign", "false")

	testFile := filepath.Join(tmpDir, "INIT")
	if err := os.WriteFile(testFile, []byte(""), 0o600); err != nil {
		t.Fatal("error writing INIT file")
	}

	runCmd(t, tmpDir, "git", "add", "INIT")
	runCmd(t, tmpDir, "git", "commit", "-m", "Initial commit")

	return tmpDir
}

// setupJJRepo creates a temporary directory and initializes a fresh jj repo.
// It returns the path to the repo and ensures local jj configs don't interfere.
func setupJJRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	runCmd(t, tmpDir, "jj", "git", "init")
	runCmd(t, tmpDir, "jj", "config", "set", "--repo", "user.name", "Test User")
	runCmd(t, tmpDir, "jj", "config", "set", "--repo", "user.email", "test@example.com")

	testFile := filepath.Join(tmpDir, "INIT")
	if err := os.WriteFile(testFile, []byte(""), 0o600); err != nil {
		t.Fatal("error writing INIT file")
	}

	runCmd(t, tmpDir, "jj", "commit", "--message", "Initial commit")

	return tmpDir
}

// runCmd is a helper to run commands inside the temp directory.
func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run %s %v: %v", name, args, err)
	}
}
