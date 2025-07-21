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
func (t Task) PriorityValue() int {
	switch t.Priority() {
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

func (t Task) MarshalTask() []byte {
	json, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func (t Task) WriteTaskJson(json []byte, p Project, kind string) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.Id(), t.Id()+".json")

		if err := os.WriteFile(file, json, 0600); err != nil {
			return WriteTaskJSONErrorMsg{err}
		}

		return WriteTaskJSONDoneMsg{Task: t, Kind: kind}
	}
}

func (t *Task) DeleteTaskFromFS(p Project) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.Id(), t.Id()+".json")

		err := os.Remove(file)
		if err != nil {
			return TaskDeleteErrorMsg{err}
		}

		return TaskDeleteDoneMsg{}
	}
}

func (t *Task) TaskToMarkdown() string {
	title := fmt.Sprintf("# %s\n\n", t.Title())

	description := fmt.Sprintf("## Description\n\n%s\n\n", t.Description())

	priority := fmt.Sprintf("## Priority\n%s\n\n", strings.ToUpper(t.Priority()))

	dueDate := ""
	if t.DueDate() != nil {
		dueDate = fmt.Sprintf("## Due at\n%s\n\n", t.DueDate())
	}

	completed := "## Done\n❌ No\n\n"
	if t.Completed() {
		completed = "## Done\n✅ Yes\n\n"
	}

	id := fmt.Sprintf("## ID\n%s", t.Id())

	return title + description + priority + dueDate + completed + id
}
