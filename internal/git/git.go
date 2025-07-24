// Copyright 2025 handlebargh
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

package git

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

type (
	GitInitDoneMsg    struct{}
	GitInitErrorMsg   struct{ Err error }
	GitCommitDoneMsg  struct{}
	GitCommitErrorMsg struct{ Err error }
	GitPullDoneMsg    struct{}
	GitPullErrorMsg   struct{ Err error }
)

func (e GitInitErrorMsg) Error() string   { return e.Err.Error() }
func (e GitCommitErrorMsg) Error() string { return e.Err.Error() }
func (e GitPullErrorMsg) Error() string   { return e.Err.Error() }

func InitCmd() tea.Cmd {
	return func() tea.Msg {
		if storage.FileExists("INIT") {
			return GitInitDoneMsg{}
		}

		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return GitInitErrorMsg{err}
		}

		if err := exec.Command("git", "init", "-b", viper.GetString("git.default_branch")).Run(); err != nil {
			return GitInitErrorMsg{err}
		}

		if err := os.WriteFile("INIT", nil, 0600); err != nil {
			return GitInitErrorMsg{err}
		}

		err := commit("INIT", "Initial commit")
		if err != nil {
			return GitInitErrorMsg{err}
		}

		return GitInitDoneMsg{}
	}
}

func CommitCmd(file, message string) tea.Cmd {
	return func() tea.Msg {
		if viper.GetBool("git.remote.enable") {
			err := pull()
			if err != nil {
				return GitPullErrorMsg{err}
			}
		}

		err := commit(file, message)
		if err != nil {
			return GitCommitErrorMsg{err}
		}

		return GitCommitDoneMsg{}
	}
}

func PullCmd() tea.Cmd {
	return func() tea.Msg {
		err := pull()
		if err != nil {
			return GitPullErrorMsg{err}
		}

		return GitPullDoneMsg{}
	}
}

func pull() error {
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return err
	}

	if err := exec.Command("git", "pull", "--rebase").Run(); err != nil {
		return err
	}

	return nil
}

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

	if viper.GetBool("git.remote.enable") {
		if err := exec.Command("git", "push", "-u", viper.GetString("git.remote.name"), viper.GetString("git.default_branch")).Run(); err != nil {
			return err
		}
	}

	return nil
}
