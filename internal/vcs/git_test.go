package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupTestRepo creates a new temporary directory, initializes a git repository,
// and sets the storage.path for viper.
func setupTestRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	err := cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "config", "commit.gpgSign", "false")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	return tempDir
}

func TestGitUser(t *testing.T) {
	setupTestRepo(t)

	user, err := gitUser()
	assert.NoError(t, err)
	assert.Equal(t, "Test User <test@example.com>", user)
}

func TestGitContributors(t *testing.T) {
	tempDir := setupTestRepo(t)

	// Create a commit to have an author
	err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0o600)
	assert.NoError(t, err)
	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	contributors, err := gitContributors()
	assert.NoError(t, err)
	assert.Contains(t, contributors, "Test User <test@example.com>")
}

func TestGitCommit(t *testing.T) {
	tempDir := setupTestRepo(t)

	filePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("hello"), 0o600)
	assert.NoError(t, err)

	output, err := gitCommit("feat: add test file", "test.txt")
	assert.NoError(t, err)
	assert.Contains(t, string(output), "feat: add test file")

	// Check that the commit was actually made
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.Dir = tempDir
	logOutput, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(logOutput), "feat: add test file")
}

func TestGitInitCmd(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)
	viper.Set("git.default_branch", "main")
	viper.Set("git.remote.enable", false)

	msg := gitInitCmd()()

	assert.IsType(t, InitDoneMsg{}, msg)

	_, err := os.Stat(filepath.Join(tempDir, "INIT"))
	assert.NoError(t, err, "INIT file should be created")
}
