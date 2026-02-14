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

// jjInitCmd initializes a jj (git compatible) repository in the configured storage path.
// It creates a jj repo with the default branch and makes an initial commit
// with a file named "INIT". If "INIT" already exists InitCmd terminates immediately.
// Returns a InitDoneMsg or InitErrorMsg.
func jjInitCmd(v *viper.Viper) tea.Cmd {
	return func() tea.Msg {
		storagePath := v.GetString("storage.path")

		root, err := os.OpenRoot(storagePath)
		if err != nil {
			return InitErrorMsg{"cannot change dir to configured storage path", err}
		}
		defer helpers.CloseWithErr(root, &err)

		if _, err := root.Stat("INIT"); err == nil {
			return InitDoneMsg{}
		}

		if !v.GetBool("jj.remote.enable") {
			var cmd *exec.Cmd
			if v.GetBool("jj.colocate") {
				cmd = exec.Command("jj", "git", "init", "--colocate")
			} else {
				cmd = exec.Command("jj", "git", "init")
			}

			cmd.Dir = storagePath

			if output, err := cmd.CombinedOutput(); err != nil {
				return InitErrorMsg{string(output), err}
			}
		}

		f, err := root.Create("INIT")
		if err != nil {
			return InitErrorMsg{"cannot create INIT file via root", err}
		}
		defer helpers.CloseWithErr(f, &err)

		if output, err := jjCommit(v, "Initial commit"); err != nil {
			return InitErrorMsg{string(output), err}
		}

		if v.GetBool("jj.remote.enable") {
			if output, err := jjPush(v); err != nil {
				return InitErrorMsg{string(output), err}
			}
		}

		return InitDoneMsg{}
	}
}

// jjCommitCmd stages and commits the specified file with the given message.
// If jj remote support is enabled, it fetches from the remote and rebases before committing.
// Returns a CommitDoneMsg or CommitErrorMsg.
func jjCommitCmd(v *viper.Viper, message string) tea.Cmd {
	return func() tea.Msg {
		if v.GetBool("jj.remote.enable") {
			if output, err := jjFetch(v); err != nil {
				return PullErrorMsg{string(output), err}
			}

			if output, err := jjRebase(v); err != nil {
				return PullErrorMsg{string(output), err}
			}
		}

		if output, err := jjCommit(v, message); err != nil {
			return CommitErrorMsg{string(output), err}
		}

		if v.GetBool("jj.remote.enable") {
			if output, err := jjPush(v); err != nil {
				return PushErrorMsg{string(output), err}
			}
		}

		return CommitDoneMsg{}
	}
}

// jjPullCmd performs a jj fetch and rebase in the configured storage path.
// Returns a PullDoneMsg or PullErrorMsg.
func jjPullCmd(v *viper.Viper) tea.Cmd {
	return func() tea.Msg {
		// Don't try to pull if repo is not initialized.
		if !storage.FileExists(v, "INIT") {
			return PullNoInitMsg{}
		}

		if output, err := jjFetch(v); err != nil {
			return PullErrorMsg{string(output), err}
		}

		if output, err := jjRebase(v); err != nil {
			return PullErrorMsg{string(output), err}
		}

		return PullDoneMsg{}
	}
}

// jjFetch changes the working directory to the configured storage path
// and performs a jj git fetch. Returns an error if any step fails.
func jjFetch(v *viper.Viper) ([]byte, error) {
	fetchCmd := exec.Command("jj", "git", "fetch")
	fetchCmd.Dir = v.GetString("storage.path")

	output, err := fetchCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// jjRebase changes the working directory to the configured storage path
// and performs a jj rebase. Returns an error if any step fails.
func jjRebase(v *viper.Viper) ([]byte, error) {
	branch := v.GetString("jj.default_branch")
	remote := v.GetString("jj.remote.name")

	rebaseCmd := exec.Command("jj", // #nosec G204 Command use validated config values
		"rebase",
		"--source",
		"@",
		"--destination", fmt.Sprintf("%s@%s", branch, remote),
	)

	rebaseCmd.Dir = v.GetString("storage.path")
	output, err := rebaseCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// jjCommit commits working copy changes with the given message.
// If remote is enabled, it pushes the commit to the configured remote and branch.
// Returns an error if any command fails.
func jjCommit(v *viper.Viper, message string) ([]byte, error) {
	storagePath := v.GetString("storage.path")

	cmd := exec.Command("jj",
		"diff",
		"--stat",
		"--revisions",
		"@-",
		"--revisions",
		"@",
	)

	cmd.Dir = storagePath
	output, err := cmd.Output()
	if err != nil {
		return output, err
	}
	if len(output) == 0 {
		return output, nil // no changes
	}

	commitCmd := exec.Command("jj", // #nosec G204 no shell interpretation
		"commit",
		"--message", message,
	)

	commitCmd.Dir = storagePath
	output, err = commitCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
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
func jjPush(v *viper.Viper) ([]byte, error) {
	storagePath := v.GetString("storage.path")
	branch := v.GetString("jj.default_branch")
	remote := v.GetString("jj.remote.name")

	bookmarkCmd := exec.Command("jj", // #nosec G204 Command uses validated config value
		"bookmark", "set", branch,
		"--revision", "@-",
	)

	bookmarkCmd.Dir = storagePath
	output, err := bookmarkCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	pushCmd := exec.Command("jj", "git", "push", // #nosec G204 Command uses validated config values
		"--allow-new",
		"--remote", remote,
		"--bookmark", branch,
	)

	pushCmd.Dir = storagePath
	output, err = pushCmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

// jjUser returns the name and email address that is returned by the
// jj config get command.
func jjUser(v *viper.Viper) (string, error) {
	storagePath := v.GetString("storage.path")

	nameCmd := exec.Command("jj", "config", "get", "user.name")
	nameCmd.Dir = storagePath
	nameOut, err := nameCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	emailCmd := exec.Command("jj", "config", "get", "user.email")
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

// jjContributorEmailAddresses returns all commit author email addresses
// found by the jj log command.
func jjContributors(v *viper.Viper) ([]string, error) {
	emailsCmd := exec.Command("jj", "log", "--template=author")
	emailsCmd.Dir = v.GetString("storage.path")

	output, err := emailsCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var authors []string
	for _, authorRaw := range strings.Split(string(output), "\n") {
		author := strings.Split(authorRaw, " ")[1:]
		authors = append(authors, strings.Join(author, " "))
	}

	return helpers.UniqueNonEmptyStrings(authors), nil
}
