package git

import (
	"os"
	"os/exec"
)

func GitInit(directory string) error {
	cmd := exec.Command("git", "init", "-b", "main", directory)
	return cmd.Run()
}

func GitPull(storageDir string) error {
	err := os.Chdir(storageDir)
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
		cmd = exec.Command("git", "push", "origin", "main")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
