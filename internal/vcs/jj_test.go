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
func setupJjTestRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

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

	return tempDir
}

func TestJjUser(t *testing.T) {
	setupJjTestRepo(t)

	user, err := jjUser()
	assert.NoError(t, err)
	assert.Equal(t, "Test User <test@example.com>", user)
}

func TestJjContributors(t *testing.T) {
	tempDir := setupJjTestRepo(t)

	// Create a commit to have an author
	err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0o600)
	assert.NoError(t, err)

	cmd := exec.Command("jj", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	contributors, err := jjContributors()
	assert.NoError(t, err)
	assert.Contains(t, contributors, "Test User <test@example.com>")
}

func TestJjCommit(t *testing.T) {
	tempDir := setupJjTestRepo(t)

	filePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("hello"), 0o600)
	assert.NoError(t, err)

	// jj automatically tracks files, so no 'add' is needed.
	// We need a first commit to be able to diff against.
	cmd := exec.Command("jj", "commit", "-m", "base")
	cmd.Dir = tempDir
	err = cmd.Run()
	assert.NoError(t, err)

	output, err := jjCommit("feat: add test file")
	assert.NoError(t, err)
	assert.Contains(t, string(output), "feat: add test file")

	// Check that the commit was actually made
	cmd = exec.Command("jj", "log", "--template=description")
	cmd.Dir = tempDir
	logOutput, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(logOutput), "feat: add test file")
}

func TestJjInitCmd(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)
	viper.Set("jj.default_branch", "main")
	viper.Set("jj.remote.enable", false)
	viper.Set("jj.colocate", true)

	msg := jjInitCmd()()

	assert.IsType(t, InitDoneMsg{}, msg)

	_, err := os.Stat(filepath.Join(tempDir, "INIT"))
	assert.NoError(t, err, "INIT file should be created")
}
