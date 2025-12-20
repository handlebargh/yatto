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

func TestResolver(t *testing.T) {
	t.Run("returns git commands when backend is git", func(t *testing.T) {
		v := viper.New()
		v.Set("vcs.backend", "git")
		assert.NotNil(t, InitCmd(v))
		assert.NotNil(t, CommitCmd(v, "test"))
		assert.NotNil(t, PullCmd(v))
	})

	t.Run("returns jj commands when backend is jj", func(t *testing.T) {
		v := viper.New()
		v.Set("vcs.backend", "jj")
		assert.NotNil(t, InitCmd(v))
		assert.NotNil(t, CommitCmd(v, "test"))
		assert.NotNil(t, PullCmd(v))
	})

	t.Run("returns nil for unknown backend", func(t *testing.T) {
		v := viper.New()
		v.Set("vcs.backend", "unknown")
		assert.Nil(t, InitCmd(v))
		assert.Nil(t, CommitCmd(v, "test"))
		assert.Nil(t, PullCmd(v))
	})

	t.Run("User function resolves correctly", func(t *testing.T) {
		// Git
		v := setupTestRepo(t)
		v.Set("vcs.backend", "git")

		user, err := User(v)
		assert.NoError(t, err)
		assert.Equal(t, "Test User <test@example.com>", user)

		// jj
		v = setupJjTestRepo(t)
		v.Set("vcs.backend", "jj")

		user, err = User(v)
		assert.NoError(t, err)
		assert.Equal(t, "Test User <test@example.com>", user)
	})

	t.Run("AllContributors function resolves correctly", func(t *testing.T) {
		// Git
		v := setupTestRepo(t)
		v.Set("vcs.backend", "git")

		makeCommit(t, v.GetString("storage.path"), "git", "Initial git commit")
		contribs, err := AllContributors(v)
		assert.NoError(t, err)
		assert.Contains(t, contribs, "Test User <test@example.com>")

		// jj
		v = setupJjTestRepo(t)
		v.Set("vcs.backend", "jj")

		makeCommit(t, v.GetString("storage.path"), "jj", "Initial jj commit")
		contribs, err = AllContributors(v)
		assert.NoError(t, err)
		assert.Contains(t, contribs, "Test User <test@example.com>")
	})
}

// makeCommit is a helper to create a commit in a repo.
func makeCommit(t *testing.T, dir, vcs, message string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o600)
	assert.NoError(t, err)

	var cmd *exec.Cmd
	if vcs == "git" {
		cmd = exec.Command("git", "add", "file.txt")
		cmd.Dir = dir
		err = cmd.Run()
		assert.NoError(t, err)
		cmd = exec.Command("git", "commit", "-m", message)
	} else {
		cmd = exec.Command("jj", "commit", "-m", message)
	}
	cmd.Dir = dir
	err = cmd.Run()
	assert.NoError(t, err)
}
