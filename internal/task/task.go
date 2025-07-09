package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/spf13/viper"
)

var uuidRegex = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`,
)

type (
	JsonWriteDoneMsg   struct{}
	JsonWriteErrorMsg  struct{ Err error }
	TaskDeleteDoneMsg  struct{}
	TaskDeleteErrorMsg struct{ Err error }
)

func (e JsonWriteErrorMsg) Error() string  { return e.Err.Error() }
func (e TaskDeleteErrorMsg) Error() string { return e.Err.Error() }

type Task struct {
	TaskId          string `json:"id"`
	TaskTitle       string `json:"title"`
	TaskDescription string `json:"description"`
	TaskPriority    string `json:"priority"`
	TaskCompleted   bool   `json:"completed"`
}

func (t Task) Id() string          { return t.TaskId }
func (t Task) Title() string       { return t.TaskTitle }
func (t Task) Description() string { return t.TaskDescription }
func (t Task) Priority() string    { return t.TaskPriority }
func (t Task) Completed() bool     { return t.TaskCompleted }
func (t Task) FilterValue() string { return t.TaskTitle }

// Function to convert priority to a numerical value for sorting.
func PriorityValue(priority string) int {
	switch priority {
	case "high":
		return 2
	case "medium":
		return 1
	case "low":
		return 0
	default:
		return -1
	}
}

func CompletedString(completed bool) string {
	if completed {
		return "done"
	}

	return "open"
}

// ReadTasksFromFS reads all tasks from the storage directory
// and returns them as a slice of Task.
func ReadTasksFromFS() []Task {
	taskFiles, err := os.ReadDir(viper.GetString("storage_dir"))
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		if entry.IsDir() || !uuidRegex.MatchString(entry.Name()) {
			continue
		}

		filePath := filepath.Join(viper.GetString("storage_dir"), entry.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		var task Task
		if err := json.Unmarshal(fileContent, &task); err != nil {
			panic(err)
		}
		tasks = append(tasks, task)
	}

	return tasks
}

func MarshalTask(uuid, title, description, priority string, completed bool) []byte {
	var task Task
	task.TaskId = uuid
	task.TaskTitle = title
	task.TaskDescription = description
	task.TaskPriority = priority
	task.TaskCompleted = completed

	json, err := json.MarshalIndent(task, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func writeJsonLogic(json []byte, task Task, message string) error {
	file := filepath.Join(viper.GetString("storage_dir"), task.Id())

	if err := os.WriteFile(file, json, 0600); err != nil {
		return err
	}

	if viper.GetBool("use_git") {
		push := viper.GetString("git_remote") != ""
		return git.GitCommitLogic(file, viper.GetString("storage_dir"), message, push)
	}

	return nil
}

func WriteJsonCmd(json []byte, task Task, message string) tea.Cmd {
	return func() tea.Msg {
		if err := writeJsonLogic(json, task, message); err != nil {
			return JsonWriteErrorMsg{Err: err}
		}
		return JsonWriteDoneMsg{}
	}
}

func deleteTaskFromFSLogic(task *Task, message string) error {
	file := filepath.Join(viper.GetString("storage_dir"), task.Id())

	err := os.Remove(file)
	if err != nil {
		panic(err)
	}

	if viper.GetBool("use_git") {
		push := viper.GetString("git_remote") == ""
		return git.GitCommitLogic(file, viper.GetString("storage_dir"), message, push)
	}

	return nil
}

func DeleteTaskFromFSCmd(task *Task, message string) tea.Cmd {
	return func() tea.Msg {
		if err := deleteTaskFromFSLogic(task, message); err != nil {
			return TaskDeleteErrorMsg{Err: err}
		}
		return TaskDeleteDoneMsg{}
	}
}
