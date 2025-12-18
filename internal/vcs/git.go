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
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

// gitInitCmd initializes a Git repository in the configured storage path.
// It creates a Git repo with the default branch and makes an initial commit
// with a file named "INIT". If "INIT" already exists InitCmd terminates immediately.
// Returns a InitDoneMsg or InitErrorMsg.
func gitInitCmd() tea.Cmd {
	return func() tea.Msg {
		storagePath := viper.GetString("storage.path")

		root, err := os.OpenRoot(storagePath)
		if err != nil {
			return InitErrorMsg{"cannot change dir to configured storage path", err}
		}
		defer root.Close()

		if _, err := root.Stat("INIT"); err == nil {
			return InitDoneMsg{}
		}

		initCmd := exec.Command("git", // #nosec G204 Command uses validated config value
			"init",
			"--initial-branch",
			viper.GetString("git.default_branch"),
		)
		initCmd.Dir = storagePath

		if output, err := initCmd.CombinedOutput(); err != nil {
			return InitErrorMsg{string(output), err}
		}

		f, err := root.Create("INIT")
		if err != nil {
			return InitErrorMsg{"cannot create INIT file via root", err}
		}
		f.Close()

		if output, err := gitCommit("Initial commit", "INIT"); err != nil {
			return InitErrorMsg{string(output), err}
		}

		if viper.GetBool("git.remote.enable") {
			if output, err := gitPush(); err != nil {
				return InitErrorMsg{string(output), err}
			}
		}

		return InitDoneMsg{}
	}
}

// gitCommitCmd stages and commits the specified files with the given message.
// If Git remote support is enabled, it pulls from the remote and rebases before pushing.
// Returns a CommitDoneMsg or CommitErrorMsg.
func gitCommitCmd(message string, files ...string) tea.Cmd {
	return func() tea.Msg {
		if output, err := gitCommit(message, files...); err != nil {
			return CommitErrorMsg{string(output), err}
		}

		if viper.GetBool("git.remote.enable") {
			if output, err := gitPull(); err != nil {
				return PullErrorMsg{string(output), err}
			}

			if output, err := gitPush(); err != nil {
				return PushErrorMsg{string(output), err}
			}
		}

		return CommitDoneMsg{}
	}
}

// gitPullCmd performs a Git pull with rebase in the configured storage path.
// Returns a PullDoneMsg or PullErrorMsg.
func gitPullCmd() tea.Cmd {
	return func() tea.Msg {
		// Don't try to pull if repo is not initialized.
		if !storage.FileExists("INIT") {
			return PullNoInitMsg{}
		}

		output, err := gitPull()
		if err != nil {
			return PullErrorMsg{string(output), err}
		}

		return PullDoneMsg{}
	}
}

// gitPull changes the working directory to the configured storage path
// and performs a git pull --rebase. Returns an error if any step fails.
func gitPull() ([]byte, error) {
	pullCmd := exec.Command("git", "pull", "--rebase")
	pullCmd.Dir = viper.GetString("storage.path")

	output, err := pullCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// gitCommit stages the specified files and commits them with the given message.
// If there are no changes, it returns nil. If remote is enabled,
// it pushes the commit to the configured remote and branch.
// Returns an error if any Git command fails.
func gitCommit(message string, files ...string) ([]byte, error) {
	storagePath := viper.GetString("storage.path")

	root, err := os.OpenRoot(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage root: %w", err)
	}
	defer root.Close()

	args := append([]string{"add"}, files...)
	addCmd := exec.Command("git", args...) // #nosec G204 Command uses only UUIDs as filenames
	addCmd.Dir = storagePath
	output, err := addCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	diffCmd := exec.Command("git",
		"diff",
		"--cached",
	)
	diffCmd.Dir = storagePath
	output, _ = diffCmd.CombinedOutput()
	if len(output) == 0 {
		return output, nil
	}

	commitCmd := exec.Command("git",
		"commit",
		"--message",
		message,
	)
	commitCmd.Dir = storagePath
	output, err = commitCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// gitPush changes the current working directory to the configured storage path
// and executes a Git push command to the specified remote and branch.
// It returns an error if changing the directory or running the Git command fails.
func gitPush() ([]byte, error) {
	pushCmd := exec.Command("git", // #nosec G204 Command uses validated config values
		"push",
		"--set-upstream",
		viper.GetString("git.remote.name"),
		viper.GetString("git.default_branch"),
	)
	pushCmd.Dir = viper.GetString("storage.path")

	output, err := pushCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// gitUser returns the name and email address that is returned by the
// git config command.
func gitUser() (string, error) {
	storagePath := viper.GetString("storage.path")

	nameCmd := exec.Command("git", "config", "user.name")
	nameCmd.Dir = storagePath
	nameOut, err := nameCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	emailCmd := exec.Command("git", "config", "user.email")
	emailCmd.Dir = storagePath

	emailOut, err := emailCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(strings.TrimSpace(string(nameOut)))
	result.WriteString(" ")
	result.WriteString(helpers.AddAngleBracketsToEmail(strings.TrimSpace(string(emailOut))))

	return result.String(), nil
}

// gitContributorEmailAddresses returns all commit author email addresses
// found by the git log command.
func gitContributors() ([]string, error) {
	emailsCmd := exec.Command("git", "log", "--format=%aN %aE")
	emailsCmd.Dir = viper.GetString("storage.path")

	output, err := emailsCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	authors := strings.Split(string(output), "\n")

	return helpers.UniqueNonEmptyStrings(authors), nil
}
