package git

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/spf13/viper"
)

const (
	branchDone       = "done"
	branchInProgress = "inProgress"
	branchOnHold     = "onHold"
)

var branches = []string{branchDone, branchInProgress, branchOnHold}

type (
	GitInitDoneMsg         struct{}
	GitInitErrorMsg        struct{ Err error }
	GitCommitDoneMsg       struct{}
	GitCommitErrorMsg      struct{ Err error }
	GitPullDoneMsg         struct{}
	GitPullErrorMsg        struct{ Err error }
	AddBranchDoneMsg       struct{ Branch items.Branch }
	AddBranchErrorMsg      struct{ Err error }
	DeleteBranchDoneMsg    struct{}
	DeleteBranchErrorMsg   struct{ Err error }
	CheckoutBranchDoneMsg  struct{}
	CheckoutBranchErrorMsg struct{ Err error }
)

func (e GitInitErrorMsg) Error() string        { return e.Err.Error() }
func (e GitCommitErrorMsg) Error() string      { return e.Err.Error() }
func (e GitPullErrorMsg) Error() string        { return e.Err.Error() }
func (e AddBranchErrorMsg) Error() string      { return e.Err.Error() }
func (e DeleteBranchErrorMsg) Error() string   { return e.Err.Error() }
func (e CheckoutBranchErrorMsg) Error() string { return e.Err.Error() }

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

		for _, branch := range branches {
			if !branchExists(branch) {
				if err := exec.Command("git", "branch", branch).Run(); err != nil {
					return GitInitErrorMsg{err}
				}
			}
		}

		return GitInitDoneMsg{}
	}
}

func CommitCmd(file, message string) tea.Cmd {
	return func() tea.Msg {
		err := commit(file, message)
		if err != nil {
			return GitCommitErrorMsg{err}
		}

		return GitCommitDoneMsg{}
	}
}

func PullCmd() tea.Cmd {
	return func() tea.Msg {
		err := Pull()
		if err != nil {
			return GitPullErrorMsg{err}
		}

		return GitPullDoneMsg{}
	}
}

func AddBranchCmd(branch items.Branch, setUpstream bool) tea.Cmd {
	return func() tea.Msg {
		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return AddBranchErrorMsg{err}
		}

		if err := exec.Command("git", "branch", branch.Title()).Run(); err != nil {
			return AddBranchErrorMsg{err}
		}

		if setUpstream {
			if err := exec.Command("git", "push", "--set-upstream", viper.GetString("git.remote"), branch.Title()).Run(); err != nil {
				return AddBranchErrorMsg{err}
			}
		}

		return AddBranchDoneMsg{Branch: branch}
	}
}

func DeleteBranchCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return DeleteBranchErrorMsg{err}
		}

		if err := exec.Command("git", "branch", "-D", branch).Run(); err != nil {
			return DeleteBranchErrorMsg{err}
		}

		return DeleteBranchDoneMsg{}
	}
}

func CheckoutBranchCmd(branch string) tea.Cmd {
	return func() tea.Msg {
		if err := os.Chdir(viper.GetString("storage.path")); err != nil {
			return CheckoutBranchErrorMsg{err}
		}

		if err := exec.Command("git", "checkout", branch).Run(); err != nil {
			return CheckoutBranchErrorMsg{err}
		}

		return CheckoutBranchDoneMsg{}
	}
}

func Pull() error {
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
		_, currentBranch, err := GetBranches()
		if err != nil {
			return err
		}

		if err := exec.Command("git", "push", "-u", viper.GetString("git.remote.name"), currentBranch).Run(); err != nil {
			return err
		}
	}

	return nil
}

func GetBranches() ([]items.Branch, string, error) {
	var branches []items.Branch
	var currentBranch string
	if err := os.Chdir(viper.GetString("storage.path")); err != nil {
		return branches, "", err
	}

	cmd := exec.Command("git", "branch", "-vv")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return branches, "", err
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		isCurrent := strings.HasPrefix(line, "*")

		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))

		branch := items.Branch{
			Name:     getBranchName(line),
			Upstream: getUpstream(line),
		}

		if isCurrent {
			currentBranch = branch.Name
		}

		branches = append(branches, branch)

	}

	if err := scanner.Err(); err != nil {
		return branches, "", err
	}

	return branches, currentBranch, nil
}

func getBranchName(line string) string {
	fields := strings.Fields(line)

	if len(fields) == 0 {
		return ""
	}

	return fields[0]
}

func getUpstream(line string) string {
	fields := strings.Fields(line)
	for _, field := range fields {
		if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
			return "Remote: " + strings.Trim(field, "[]")
		}
	}
	return "local only"
}

func branchExists(branch string) bool {
	if err := exec.Command("git", "rev-parse", "--verify", branch).Run(); err == nil {
		// Branch exists
		return true
	}

	return false
}
