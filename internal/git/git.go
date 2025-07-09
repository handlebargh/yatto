package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

type (
	GitDoneMsg  struct{}
	GitErrorMsg struct{ Err error }
)

func GitInit(storageDir string) tea.Cmd {
	return func() tea.Msg {
		if err := gitInitLogic(storageDir); err != nil {
			return GitErrorMsg{err}
		}
		return GitDoneMsg{}
	}
}

func GitCommit(file, storageDir, message string, push bool) tea.Cmd {
	return func() tea.Msg {
		if err := GitCommitLogic(file, storageDir, message, push); err != nil {
			return GitErrorMsg{err}
		}
		return GitDoneMsg{}
	}
}

func GitPull(storageDir string) tea.Cmd {
	return func() tea.Msg {
		if err := GitPullLogic(storageDir); err != nil {
			return GitErrorMsg{err}
		}
		return GitDoneMsg{}
	}
}

func gitInitLogic(storageDir string) error {
	if fileExists(storageDir, "INIT") {
		return nil
	}

	cmd := exec.Command("git", "init", "-b", "main", storageDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	initFilePath := filepath.Join(storageDir, "INIT")
	if err := os.WriteFile(initFilePath, nil, 0600); err != nil {
		return err
	}

	if viper.GetString("git_remote") == "" {
		return GitCommitLogic("INIT", storageDir, "Initial commit", false)
	}

	if err := ensureGitRemote(storageDir, "origin", viper.GetString("git_remote")); err != nil {
		return err
	}

	return GitCommitLogic("INIT", storageDir, "Initial commit", true)
}

func GitCommitLogic(file, storageDir, message string, push bool) error {
	if err := os.Chdir(storageDir); err != nil {
		return err
	}

	if err := exec.Command("git", "add", file).Run(); err != nil {
		return err
	}

	if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
		return err
	}

	if push {
		if err := exec.Command("git", "push", "-u", "origin", "main").Run(); err != nil {
			return err
		}
	}

	return nil
}

func GitPullLogic(storageDir string) error {
	if err := os.Chdir(storageDir); err != nil {
		return err
	}

	if err := ensureGitRemote(storageDir, "origin", viper.GetString("git_remote")); err != nil {
		return err
	}

	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		return err
	}

	return exec.Command("git", "rebase", "origin/main").Run()
}

func fileExists(storageDir, filename string) bool {
	info, err := os.Stat(filepath.Join(storageDir, filename))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func ensureGitRemote(storageDir, remoteName, remoteURL string) error {
	err := os.Chdir(storageDir)
	if err != nil {
		return err
	}

	// Check if the remote exists
	cmdCheck := exec.Command("git", "remote", "get-url", remoteName)
	var out bytes.Buffer
	cmdCheck.Stdout = &out
	err = cmdCheck.Run()

	if err == nil {
		// Remote exists - update its URL
		existingURL := strings.TrimSpace(out.String())
		if existingURL != remoteURL {
			cmdSet := exec.Command("git", "remote", "set-url", remoteName, remoteURL)
			if err := cmdSet.Run(); err != nil {
				return err
			}
		}
	} else {
		// Remote doesn't exist - add it
		cmdAdd := exec.Command("git", "remote", "add", remoteName, remoteURL)
		if err := cmdAdd.Run(); err != nil {
			return err
		}
	}

	return nil
}
