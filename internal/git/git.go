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
	"errors"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

// ErrorNoInit is returned when a git pull command is executed
// in a non-initialized storage repository.
var ErrorNoInit = errors.New(
	"trying to pull but local repo is not initialized.\nPlease disable git.remote and try again",
)

type (
	// InitDoneMsg is returned when Git initialization completes successfully.
	InitDoneMsg struct{}

	// InitErrorMsg is returned when Git initialization fails.
	InitErrorMsg struct{ Err error }

	// CommitDoneMsg is returned when a Git commit completes successfully.
	CommitDoneMsg struct{}

	// CommitErrorMsg is returned when a Git commit fails.
	CommitErrorMsg struct{ Err error }

	// PullDoneMsg is returned when a Git pull operation completes successfully.
	PullDoneMsg struct{}

	// PullErrorMsg is returned when a Git pull operation fails.
	PullErrorMsg struct{ Err error }

	// PullNoInitMsg is returned when a Git pull operation didn't run
	// because the repository's INIT file is missing.
	PullNoInitMsg struct{}

	// PushErrorMsg is returned when a Git push operation fails.
	PushErrorMsg struct{ Err error }
)

// Error implements the error interface for GitInitErrorMsg.
func (e InitErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for GitCommitErrorMsg.
func (e CommitErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for GitPullErrorMsg.
func (e PullErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for GitPushErrorMsg.
func (e PushErrorMsg) Error() string { return e.Err.Error() }

// InitCmd initializes a Git repository in the configured storage path.
// It creates a Git repo with the default branch and makes an initial commit
// with a file named "INIT". If "INIT" already exists InitCmd terminates immediately.
// Returns a GitInitDoneMsg or GitInitErrorMsg.
func InitCmd() tea.Cmd {
	return func() tea.Msg {
		if storage.FileExists("INIT") {
			return InitDoneMsg{}
		}

		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return InitErrorMsg{err}
		}

		if err := exec.Command("git", "init", "-b",
			viper.GetString("git.default_branch")).Run(); err != nil {
			return InitErrorMsg{err}
		}

		if err := os.WriteFile("INIT", nil, 0o600); err != nil {
			return InitErrorMsg{err}
		}

		err := commit("INIT", "Initial commit")
		if err != nil {
			return InitErrorMsg{err}
		}

		return InitDoneMsg{}
	}
}

// CommitCmd stages and commits the specified file with the given message.
// If Git remote support is enabled, it pulls from the remote before committing.
// Returns a GitCommitDoneMsg or GitCommitErrorMsg.
func CommitCmd(file, message string) tea.Cmd {
	return func() tea.Msg {
		if err := commit(file, message); err != nil {
			return CommitErrorMsg{err}
		}

		if viper.GetBool("git.remote.enable") {
			if err := pull(); err != nil {
				return PullErrorMsg{err}
			}

			if err := push(); err != nil {
				return PushErrorMsg{err}
			}
		}

		return CommitDoneMsg{}
	}
}

// PullCmd performs a Git pull with rebase in the configured storage path.
// Returns a GitPullDoneMsg or GitPullErrorMsg.
func PullCmd() tea.Cmd {
	return func() tea.Msg {
		// Don't try to pull if repo is not initialized.
		if !storage.FileExists("INIT") {
			return PullNoInitMsg{}
		}

		err := pull()
		if err != nil {
			return PullErrorMsg{err}
		}

		return PullDoneMsg{}
	}
}

// pull changes the working directory to the configured storage path
// and performs a `git pull --rebase`. Returns an error if any step fails.
func pull() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("git", "pull", "--rebase").Run(); err != nil {
		return err
	}

	return nil
}

// commit stages the specified file and commits it with the given message.
// If there are no changes, it returns nil. If remote Git is enabled,
// it pushes the commit to the configured remote and branch.
// Returns an error if any Git command fails.
func commit(file, message string) error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("git", "add", file).Run(); err != nil {
		return err
	}

	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err == nil {
		// Exit code 0 = no staged changes
		return nil // Already committed.
	}

	if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
		return err
	}

	return nil
}

// push changes the current working directory to the configured storage path
// and executes a Git push command to the specified remote and branch.
// It returns an error if changing the directory or running the Git command fails.
func push() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("git", "push", "-u",
		viper.GetString("git.remote.name"),
		viper.GetString("git.default_branch")).Run(); err != nil {
		return err
	}

	return nil
}
