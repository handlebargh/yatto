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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

// jjInitCmd initializes a jj (git compatible) repository in the configured storage path.
// It creates a jj repo with the default branch and makes an initial commit
// with a file named "INIT". If "INIT" already exists InitCmd terminates immediately.
// Returns a InitDoneMsg or InitErrorMsg.
func jjInitCmd() tea.Cmd {
	return func() tea.Msg {
		if storage.FileExists("INIT") {
			return InitDoneMsg{}
		}

		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return InitErrorMsg{err}
		}

		var cmd *exec.Cmd
		if viper.GetBool("jj.colocate") {
			cmd = exec.Command("jj", "git", "init", "--colocate")
		} else {
			cmd = exec.Command("jj", "git", "init")
		}

		if err := cmd.Run(); err != nil {
			return InitErrorMsg{err}
		}

		if err := os.WriteFile("INIT", nil, 0o600); err != nil {
			return InitErrorMsg{err}
		}

		err := jjCommit("Initial commit")
		if err != nil {
			return InitErrorMsg{err}
		}

		return InitDoneMsg{}
	}
}

// jjCommitCmd stages and commits the specified file with the given message.
// If jj remote support is enabled, it fetches from the remote before committing.
// Returns a CommitDoneMsg or CommitErrorMsg.
func jjCommitCmd(message string) tea.Cmd {
	return func() tea.Msg {
		if err := jjCommit(message); err != nil {
			return CommitErrorMsg{err}
		}

		if viper.GetBool("jj.remote.enable") {
			if err := jjFetch(); err != nil {
				return PullErrorMsg{err}
			}

			if err := jjPush(); err != nil {
				return PushErrorMsg{err}
			}
		}

		return CommitDoneMsg{}
	}
}

// jjPullCmd performs a jj fetch and rebase in the configured storage path.
// Returns a PullDoneMsg or PullErrorMsg.
func jjPullCmd() tea.Cmd {
	return func() tea.Msg {
		// Don't try to pull if repo is not initialized.
		if !storage.FileExists("INIT") {
			return PullNoInitMsg{}
		}

		if err := jjFetch(); err != nil {
			return PullErrorMsg{err}
		}

		if err := jjRebase(); err != nil {
			return PullErrorMsg{err}
		}

		return PullDoneMsg{}
	}
}

// jjFetch changes the working directory to the configured storage path
// and performs a jj git fetch. Returns an error if any step fails.
func jjFetch() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("jj", "git", "fetch").Run(); err != nil {
		return err
	}

	return nil
}

// jjRebase changes the working directory to the configured storage path
// and performs a jj rebase. Returns an error if any step fails.
func jjRebase() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("jj", "rebase", "-s", "@", "-d", "trunk()").Run(); err != nil {
		return err
	}

	return nil
}

// jjCommit commits working copy changes with the given message.
// If remote is enabled, it pushes the commit to the configured remote and branch.
// Returns an error if any Git command fails.
func jjCommit(message string) error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	cmd := exec.Command("jj", "diff", "--stat", "-r", "@-", "-r", "@")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return nil // no changes
	}

	if err := exec.Command("jj", "commit", "-m", message).Run(); err != nil {
		return err
	}

	return nil
}

// jjPush updates the default branch bookmark in the local Jujutsu repository
// and pushes it to the configured remote.
//
// The function performs the following steps:
//  1. Changes the working directory to the configured storage path.
//  2. Moves the default branch bookmark (from config key "jj.default_branch")
//     to point to @-, i.e. the parent of the working copy commit.
//  3. Pushes that bookmark to the Git remote specified in
//     "jj.remote.name".
func jjPush() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	defaultBranch := viper.GetString("jj.default_branch")
	if err := exec.Command("jj", "bookmark", "set",
		defaultBranch, "--revision=@-", defaultBranch).Run(); err != nil {
		return err
	}

	if err := exec.Command("jj", "git", "push",
		"--allow-new",
		"--remote",
		viper.GetString("jj.remote.name"),
		"--bookmark",
		viper.GetString("jj.default_branch")).Run(); err != nil {
		return err
	}

	return nil
}
