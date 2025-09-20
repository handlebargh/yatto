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

// Package git provides internal helpers for managing Git operations
// such as initialization, committing, and pulling in the configured
// storage directory.
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary directory and configures viper to use it.
func setupTestRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Store original config values
	originalStoragePath := viper.GetString("storage.path")
	originalDefaultBranch := viper.GetString("git.default_branch")
	originalRemoteEnable := viper.GetBool("git.remote.enable")
	originalRemoteName := viper.GetString("git.remote.name")

	// Set up isolated Git environment for all platforms
	setupGitEnvironment(t, tmpDir)

	// Set test config
	viper.Set("storage.path", tmpDir)
	viper.Set("git.default_branch", "main")
	viper.Set("git.remote.enable", false)
	viper.Set("git.remote.name", "origin")

	// Restore config after test
	t.Cleanup(func() {
		viper.Set("storage.path", originalStoragePath)
		viper.Set("git.default_branch", originalDefaultBranch)
		viper.Set("git.remote.enable", originalRemoteEnable)
		viper.Set("git.remote.name", originalRemoteName)
	})
}

// setupGitEnvironment configures Git environment for testing across platforms.
func setupGitEnvironment(t *testing.T, tmpDir string) {
	// Basic Git configuration via environment variables
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	// Disable system and global config
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", getNullDevice())
	t.Setenv("GIT_CONFIG_SYSTEM", getNullDevice())

	// Set home directory based on platform
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
		t.Setenv("HOMEDRIVE", "")
		t.Setenv("HOMEPATH", "")
	} else {
		t.Setenv("HOME", tmpDir)
	}

	// Additional isolation
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Disable GPG signing and other interactive features
	t.Setenv("GIT_CONFIG", filepath.Join(tmpDir, ".gitconfig"))

	// Create a minimal .gitconfig in the test directory
	gitConfig := filepath.Join(tmpDir, ".gitconfig")
	gitConfigContent := `[user]
	name = Test User
	email = test@example.com
[commit]
	gpgsign = false
[init]
	defaultBranch = main
[pull]
	rebase = true
[core]
	autocrlf = false
[advice]
	detachedHead = false
`
	require.NoError(t, os.WriteFile(gitConfig, []byte(gitConfigContent), 0o644))
}

// getNullDevice returns the null device path for the current platform.
func getNullDevice() string {
	if runtime.GOOS == "windows" {
		return "NUL"
	}
	return "/dev/null"
}

// gitLogOneline returns the oneline git log for testing.
func gitLogOneline(_ *testing.T, dir string) []string {
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return []string{} // Empty repo
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}
	return lines
}

// hasCommitWithMessage checks if a commit with the given message exists.
func hasCommitWithMessage(t *testing.T, dir, message string) bool {
	logs := gitLogOneline(t, dir)
	for _, log := range logs {
		if strings.Contains(log, message) {
			return true
		}
	}
	return false
}

func TestInitCmd_Success(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Execute InitCmd
	cmd := InitCmd()
	msg := cmd()

	// Should return success message
	assert.IsType(t, InitDoneMsg{}, msg)

	// Should create INIT file
	initPath := filepath.Join(tmpDir, "INIT")
	assert.FileExists(t, initPath)

	// Should create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	assert.DirExists(t, gitDir)

	// Should have initial commit
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Initial commit"))

	// Should be on the correct branch
	gitCmd := exec.Command("git", "branch", "--show-current")
	gitCmd.Dir = tmpDir
	output, err := gitCmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "main", strings.TrimSpace(string(output)))
}

func TestInitCmd_AlreadyInitialized(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Create INIT file manually
	initPath := filepath.Join(tmpDir, "INIT")
	err := os.WriteFile(initPath, nil, 0o600)
	require.NoError(t, err)

	// Execute InitCmd
	cmd := InitCmd()
	msg := cmd()

	// Should return success message immediately
	assert.IsType(t, InitDoneMsg{}, msg)

	// Should not create .git directory since we didn't actually initialize
	gitDir := filepath.Join(tmpDir, ".git")
	assert.NoDirExists(t, gitDir)
}

func TestInitCmd_GitInitFails(t *testing.T) {
	setupTestRepo(t)

	// Use a path that doesn't exist and can't be created
	viper.Set("storage.path", "/nonexistent/path/that/cannot/be/created")

	t.Cleanup(func() {
		viper.Set("storage.path", "")
	})

	// Execute InitCmd
	cmd := InitCmd()
	msg := cmd()

	// Should return error message
	//goland:noinspection GoTypeAssertionOnErrors
	errorMsg, ok := msg.(InitErrorMsg)
	assert.True(t, ok)
	assert.Error(t, errorMsg.Err)
}

func TestCommitCmd_Success(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// Create a test file
	testFile := "test.txt"
	testPath := filepath.Join(tmpDir, testFile)
	err := os.WriteFile(testPath, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Execute CommitCmd
	commitCmd := CommitCmd(testFile, "Add test file")
	msg := commitCmd()

	// Should return success message
	assert.IsType(t, CommitDoneMsg{}, msg)

	// Should have the commit
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Add test file"))

	// Should have 2 commits total (initial + test file)
	logs := gitLogOneline(t, tmpDir)
	assert.Len(t, logs, 2)
}

func TestCommitCmd_NoChanges(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// Try to commit INIT file again (no changes)
	commitCmd := CommitCmd("INIT", "Try to commit again")
	msg := commitCmd()

	// Should still return success (no error for no changes)
	assert.IsType(t, CommitDoneMsg{}, msg)

	// Should only have 1 commit (the initial one)
	logs := gitLogOneline(t, tmpDir)
	assert.Len(t, logs, 1)
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Initial commit"))
	assert.False(t, hasCommitWithMessage(t, tmpDir, "Try to commit again"))
}

func TestCommitCmd_FileNotExists(t *testing.T) {
	setupTestRepo(t)

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// Try to commit non-existent file
	commitCmd := CommitCmd("nonexistent.txt", "Add nonexistent file")
	msg := commitCmd()

	// Should return error message
	//goland:noinspection GoTypeAssertionOnErrors
	errorMsg, ok := msg.(CommitErrorMsg)
	assert.True(t, ok)
	assert.Error(t, errorMsg.Err)
}

func TestPullCmd_Success(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// For this test, we need a remote. Let's create a bare repo to simulate it.
	bareDir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = bareDir
	require.NoError(t, cmd.Run())

	// Add the bare repo as remote
	cmd = exec.Command("git", "remote", "add", "origin", bareDir)
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Push initial commit to have something to pull
	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Now test pull
	pullCmd := PullCmd()
	msg := pullCmd()

	// Should return success message
	assert.IsType(t, PullDoneMsg{}, msg)
}

func TestPullCmd_NoInit(t *testing.T) {
	setupTestRepo(t)

	// Don't initialize repo (no INIT file)
	// Execute PullCmd
	pullCmd := PullCmd()
	msg := pullCmd()

	// Should return no init message
	assert.IsType(t, PullNoInitMsg{}, msg)
}

func TestPullCmd_NoRemote(t *testing.T) {
	setupTestRepo(t)

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// Execute PullCmd without setting up remote
	pullCmd := PullCmd()
	msg := pullCmd()

	// Should return error message
	//goland:noinspection GoTypeAssertionOnErrors
	errorMsg, ok := msg.(PullErrorMsg)
	assert.True(t, ok)
	assert.Error(t, errorMsg.Err)
}

func TestCommitCmd_WithRemoteEnabled(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Enable remote
	viper.Set("git.remote.enable", true)

	// Initialize repo first
	initCmd := InitCmd()
	initMsg := initCmd()
	require.IsType(t, InitDoneMsg{}, initMsg)

	// Create a test file
	testFile := "test.txt"
	testPath := filepath.Join(tmpDir, testFile)
	err := os.WriteFile(testPath, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Execute CommitCmd (this will fail because no remote is set up)
	commitCmd := CommitCmd(testFile, "Add test file")
	msg := commitCmd()

	// Should return error because pull will fail
	//goland:noinspection GoTypeAssertionOnErrors
	errorMsg, ok := msg.(PullErrorMsg)
	assert.True(t, ok)
	assert.Error(t, errorMsg.Err)
}

// Integration test that tests the full workflow
func TestFullGitWorkflow(t *testing.T) {
	setupTestRepo(t)
	tmpDir := viper.GetString("storage.path")

	// Step 1: Initialize
	initCmd := InitCmd()
	msg := initCmd()
	require.IsType(t, InitDoneMsg{}, msg)

	// Verify initialization
	assert.FileExists(t, filepath.Join(tmpDir, "INIT"))
	assert.DirExists(t, filepath.Join(tmpDir, ".git"))

	// Step 2: Add and commit a file
	testFile := "tasks.json"
	testPath := filepath.Join(tmpDir, testFile)
	testContent := `[{"id": 1, "text": "Test task", "completed": false}]`
	err := os.WriteFile(testPath, []byte(testContent), 0o644)
	require.NoError(t, err)

	commitCmd := CommitCmd(testFile, "Add tasks file")
	msg = commitCmd()
	require.IsType(t, CommitDoneMsg{}, msg)

	// Step 3: Modify and commit again
	updatedContent := `[{"id": 1, "text": "Updated task", "completed": true}]`
	err = os.WriteFile(testPath, []byte(updatedContent), 0o644)
	require.NoError(t, err)

	commitCmd = CommitCmd(testFile, "Update task status")
	msg = commitCmd()
	require.IsType(t, CommitDoneMsg{}, msg)

	// Verify final state
	logs := gitLogOneline(t, tmpDir)
	assert.Len(t, logs, 3) // initial + add tasks + update task
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Initial commit"))
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Add tasks file"))
	assert.True(t, hasCommitWithMessage(t, tmpDir, "Update task status"))

	// Verify file content
	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, updatedContent, string(content))
}
