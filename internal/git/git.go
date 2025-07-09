package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func GitInit(storageDir string) error {
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
		err := GitCommit("INIT", storageDir, "Initial commit", false)
		if err != nil {
			return err
		}
	} else {
		err := ensureGitRemote(storageDir, "origin", viper.GetString("git_remote"))
		if err != nil {
			return err
		}

		err = GitCommit("INIT", storageDir, "Initial commit", true)
		if err != nil {
			return err
		}
	}

	return nil
}

func GitPull(storageDir string) error {
	err := os.Chdir(storageDir)
	if err != nil {
		return err
	}

	err = ensureGitRemote(storageDir, "origin", viper.GetString("git_remote"))
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "fetch", "origin")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "rebase", "origin/main")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func GitCommit(file, storageDir, message string, push bool) error {
	err := os.Chdir(storageDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "add", file)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return err
	}

	if push {
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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
