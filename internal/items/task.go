package items

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

const DueDateLayout = "2006-01-02 15:04:05"

var uuidRegex = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}\.json$`,
)

type (
	WriteTaskJSONDoneMsg struct {
		Task Task
		Kind string
	}
	WriteTaskJSONErrorMsg struct{ Err error }
	TaskDeleteDoneMsg     struct{}
	TaskDeleteErrorMsg    struct{ Err error }
)

func (e WriteTaskJSONErrorMsg) Error() string { return e.Err.Error() }
func (e TaskDeleteErrorMsg) Error() string    { return e.Err.Error() }

type Task struct {
	TaskId          string     `json:"id"`
	TaskTitle       string     `json:"title"`
	TaskDescription string     `json:"description,omitempty"`
	TaskPriority    string     `json:"priority"`
	TaskDueDate     *time.Time `json:"due_date,omitempty"`
	TaskCompleted   bool       `json:"completed"`
}

func (t Task) Id() string                         { return t.TaskId }
func (t *Task) SetId(id string)                   { t.TaskId = id }
func (t Task) Title() string                      { return t.TaskTitle }
func (t *Task) SetTitle(title string)             { t.TaskTitle = title }
func (t Task) Description() string                { return t.TaskDescription }
func (t *Task) SetDescription(description string) { t.TaskDescription = description }
func (t Task) Priority() string                   { return t.TaskPriority }
func (t *Task) SetPriority(priority string)       { t.TaskPriority = priority }
func (t Task) DueDate() *time.Time                { return t.TaskDueDate }
func (t *Task) SetDueDate(dueDate *time.Time)     { t.TaskDueDate = dueDate }
func (t Task) Completed() bool                    { return t.TaskCompleted }
func (t *Task) SetCompleted(completed bool)       { t.TaskCompleted = completed }
func (t Task) FilterValue() string                { return t.TaskTitle }

// Converts a time.Time object to string.
func (t Task) DueDateToString() string {
	if t.TaskDueDate != nil {
		return t.DueDate().Format(DueDateLayout)
	}

	return ""
}

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
func ReadTasksFromFS(project *Project) []Task {
	storageDir := viper.GetString("storage.path")
	taskFiles, err := os.ReadDir(filepath.Join(storageDir, project.Id()))
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		if entry.IsDir() || !uuidRegex.MatchString(entry.Name()) {
			continue
		}

		filePath := filepath.Join(storageDir, project.Id(), entry.Name())
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

func MarshalTask(uuid, title, description, priority string, dueDate *time.Time, completed bool) []byte {
	var task Task
	task.SetId(uuid)
	task.SetTitle(title)
	task.SetDescription(description)
	task.SetPriority(priority)
	task.SetDueDate(dueDate)
	task.SetCompleted(completed)

	json, err := json.MarshalIndent(task, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func WriteTaskJson(json []byte, project Project, task Task, kind string) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), project.Id(), task.Id()+".json")

		if err := os.WriteFile(file, json, 0600); err != nil {
			return WriteTaskJSONErrorMsg{err}
		}

		return WriteTaskJSONDoneMsg{Task: task, Kind: kind}
	}
}

func DeleteTaskFromFS(project Project, task *Task) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), project.Id(), task.Id()+".json")

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

	dueDate := ""
	if task.DueDate() != nil {
		dueDate = fmt.Sprintf("## Due at\n%s\n\n", task.DueDate())
	}

	completed := "## Done\n❌ No\n\n"
	if task.Completed() {
		completed = "## Done\n✅ Yes\n\n"
	}

	id := fmt.Sprintf("## ID\n%s", task.Id())

	return title + description + priority + dueDate + completed + id
}
