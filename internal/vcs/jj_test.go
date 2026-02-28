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

package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupJjTestRepo creates a new temporary directory, initializes a jj repository,
// and sets the storage.path for viper.
func setupJjTestRepo(t *testing.T) *viper.Viper {
	t.Helper()

	tempDir := t.TempDir()
	v := viper.New()
	v.Set("storage.path", tempDir)

	cmd := exec.Command("jj", "git", "init", "--colocate")
	cmd.Dir = tempDir
	err := cmd.Run()
	assert.NoError(t, err)

	// jj requires a username and email to be set
	cmd = exec.Command("jj", "config", "set", "--user", "user.name", "Test User")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("jj", "config", "set", "--user", "user.email", "test@example.com")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	return v
}

func TestJjUser(t *testing.T) {
	v := setupJjTestRepo(t)

	user, err := jjUser(v)
	assert.NoError(t, err)
	assert.Equal(t, "Test User <test@example.com>", user)
}

func TestJjContributors(t *testing.T) {
	v := setupJjTestRepo(t)
	storagePath := v.GetString("storage.path")

	// Create a commit to have an author
	err := os.WriteFile(filepath.Join(storagePath, "file.txt"), []byte("content"), 0o600)
	assert.NoError(t, err)

	cmd := exec.Command("jj", "commit", "-m", "Initial commit")
	cmd.Dir = storagePath
	err = cmd.Run()
	assert.NoError(t, err)

	contributors, err := jjContributors(v)
	assert.NoError(t, err)
	assert.Contains(t, contributors, "Test User <test@example.com>")
}

func TestJjCommit(t *testing.T) {
	v := setupJjTestRepo(t)
	storagePath := v.GetString("storage.path")

	filePath := filepath.Join(storagePath, "test.txt")
	err := os.WriteFile(filePath, []byte("hello"), 0o600)
	assert.NoError(t, err)

	// jj automatically tracks files, so no 'add' is needed.
	// We need a first commit to be able to diff against.
	cmd := exec.Command("jj", "commit", "-m", "base")
	cmd.Dir = storagePath
	err = cmd.Run()
	assert.NoError(t, err)

	output, err := jjCommit(v, "feat: add test file")
	assert.NoError(t, err)
	assert.Contains(t, string(output), "feat: add test file")

	// Check that the commit was actually made
	cmd = exec.Command("jj", "log", "--template=description")
	cmd.Dir = storagePath
	logOutput, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(logOutput), "feat: add test file")
}

func TestJjInitCmd(t *testing.T) {
	tempDir := t.TempDir()
	v := viper.New()
	v.Set("storage.path", tempDir)
	v.Set("jj.default_branch", "main")
	v.Set("jj.remote.enable", false)
	v.Set("jj.colocate", true)

	msg := jjInitCmd(v)()

	assert.IsType(t, InitDoneMsg{}, msg)

	_, err := os.Stat(filepath.Join(tempDir, "INIT"))
	assert.NoError(t, err, "INIT file should be created")
}
