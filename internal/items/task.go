package items

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

var uuidRegex = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`,
)

type (
	WriteJSONDoneMsg struct {
		Task Task
		Kind string
	}
	WriteJSONErrorMsg  struct{ Err error }
	TaskDeleteDoneMsg  struct{}
	TaskDeleteErrorMsg struct{ Err error }
)

func (e WriteJSONErrorMsg) Error() string  { return e.Err.Error() }
func (e TaskDeleteErrorMsg) Error() string { return e.Err.Error() }

type Task struct {
	TaskId          string `json:"id"`
	TaskTitle       string `json:"title"`
	TaskDescription string `json:"description,omitempty"`
	TaskPriority    string `json:"priority"`
	TaskCompleted   bool   `json:"completed"`
}

func (t Task) Id() string                         { return t.TaskId }
func (t *Task) SetId(id string)                   { t.TaskId = id }
func (t Task) Title() string                      { return t.TaskTitle }
func (t *Task) SetTitle(title string)             { t.TaskTitle = title }
func (t Task) Description() string                { return t.TaskDescription }
func (t *Task) SetDescription(description string) { t.TaskDescription = description }
func (t Task) Priority() string                   { return t.TaskPriority }
func (t *Task) SetPriority(priority string)       { t.TaskPriority = priority }
func (t Task) Completed() bool                    { return t.TaskCompleted }
func (t *Task) SetCompleted(completed bool)       { t.TaskCompleted = completed }
func (t Task) FilterValue() string                { return t.TaskTitle }

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
	storageDir := viper.GetString("storage.path")
	taskFiles, err := os.ReadDir(storageDir)
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		if entry.IsDir() || !uuidRegex.MatchString(entry.Name()) {
			continue
		}

		filePath := filepath.Join(storageDir, entry.Name())
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
	task.SetId(uuid)
	task.SetTitle(title)
	task.SetDescription(description)
	task.SetPriority(priority)
	task.SetCompleted(completed)

	json, err := json.MarshalIndent(task, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func WriteJson(json []byte, task Task, kind string) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), task.Id())

		if err := os.WriteFile(file, json, 0600); err != nil {
			return WriteJSONErrorMsg{err}
		}

		return WriteJSONDoneMsg{Task: task, Kind: kind}
	}
}

func DeleteTaskFromFS(task *Task) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), task.Id())

		err := os.Remove(file)
		if err != nil {
			return TaskDeleteErrorMsg{err}
		}

		return TaskDeleteDoneMsg{}
	}
}

func TaskToMarkdown(task *Task) string {
	title := fmt.Sprintf("# %s\n\n", task.Title())

	description := fmt.Sprintf("## Description\n\n%s\n\n", task.Description())

	priority := fmt.Sprintf("## Priority\n%s\n\n", strings.ToUpper(task.Priority()))

	completed := "## Done\n❌ No\n\n"
	if task.Completed() {
		completed = "## Done\n✅ Yes\n\n"
	}

	id := fmt.Sprintf("## ID\n%s", task.Id())

	return title + description + priority + completed + id
}
