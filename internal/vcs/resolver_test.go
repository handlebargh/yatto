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
		viper.Set("vcs.backend", "git")
		assert.NotNil(t, InitCmd())
		assert.NotNil(t, CommitCmd("test"))
		assert.NotNil(t, PullCmd())
	})

	t.Run("returns jj commands when backend is jj", func(t *testing.T) {
		viper.Set("vcs.backend", "jj")
		assert.NotNil(t, InitCmd())
		assert.NotNil(t, CommitCmd("test"))
		assert.NotNil(t, PullCmd())
	})

	t.Run("returns nil for unknown backend", func(t *testing.T) {
		viper.Set("vcs.backend", "unknown")
		assert.Nil(t, InitCmd())
		assert.Nil(t, CommitCmd("test"))
		assert.Nil(t, PullCmd())
	})

	t.Run("User function resolves correctly", func(t *testing.T) {
		// Git
		viper.Set("vcs.backend", "git")
		setupTestRepo(t)
		user, err := User()
		assert.NoError(t, err)
		assert.Equal(t, "Test User <test@example.com>", user)

		// Jj
		viper.Set("vcs.backend", "jj")
		setupJjTestRepo(t)
		user, err = User()
		assert.NoError(t, err)
		assert.Equal(t, "Test User <test@example.com>", user)
	})

	t.Run("AllContributors function resolves correctly", func(t *testing.T) {
		// Git
		viper.Set("vcs.backend", "git")
		gitDir := setupTestRepo(t)
		makeCommit(t, gitDir, "git", "Initial git commit")
		contribs, err := AllContributors()
		assert.NoError(t, err)
		assert.Contains(t, contribs, "Test User <test@example.com>")

		// Jj
		viper.Set("vcs.backend", "jj")
		jjDir := setupJjTestRepo(t)
		makeCommit(t, jjDir, "jj", "Initial jj commit")
		contribs, err = AllContributors()
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
