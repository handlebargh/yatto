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

// Package items provides internal types and utilities for managing task and project items,
// including creation, serialization, deletion, and formatting.
// Projects are stored as directories, each containing a JSON file with project metadata
// and multiple task files.
// Tasks are stored as JSON files and support basic metadata like priority, labels, and due dates.
package items

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

const ellipses = "..."

type (
	// WriteTaskJSONDoneMsg indicates successful write of a Task JSON file.
	WriteTaskJSONDoneMsg struct {
		Task Task
		Kind string
	}

	// WriteTaskJSONErrorMsg is returned when a Task fails to serialize or write to disk.
	WriteTaskJSONErrorMsg struct{ Err error }

	// TaskDeleteDoneMsg indicates successful deletion of a Task from disk.
	TaskDeleteDoneMsg struct{}

	// TaskDeleteErrorMsg is returned when a Task fails to delete from disk.
	TaskDeleteErrorMsg struct{ Err error }
)

// Error implements the error interface for WriteTaskJSONErrorMsg.
func (e WriteTaskJSONErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for TaskDeleteErrorMsg.
func (e TaskDeleteErrorMsg) Error() string { return e.Err.Error() }

// Task represents a to-do item with metadata like title, due date, priority,
// and labels. Tasks are serialized to and from JSON files in storage.
type Task struct {
	TaskID          string     `json:"id"`
	TaskTitle       string     `json:"title"`
	TaskDescription string     `json:"description,omitempty"`
	TaskPriority    string     `json:"priority"`
	TaskDueDate     *time.Time `json:"due_date,omitempty"`
	TaskLabels      string     `json:"labels,omitempty"`
	TaskInProgress  bool       `json:"in_progress"`
	TaskCompleted   bool       `json:"completed"`
}

// ID returns the task's ID.
func (t Task) ID() string { return t.TaskID }

// SetID sets the task's ID.
func (t *Task) SetID(id string) { t.TaskID = id }

// Title returns the task's title.
func (t Task) Title() string { return t.TaskTitle }

// SetTitle sets the task's title.
func (t *Task) SetTitle(title string) { t.TaskTitle = title }

// Description returns the task's description.
func (t Task) Description() string { return t.TaskDescription }

// SetDescription sets the task's description.
func (t *Task) SetDescription(description string) { t.TaskDescription = description }

// Priority returns the task's priority.
func (t Task) Priority() string { return t.TaskPriority }

// SetPriority sets the task's priority.
func (t *Task) SetPriority(priority string) { t.TaskPriority = priority }

// DueDate returns the task's due date, if any.
func (t Task) DueDate() *time.Time { return t.TaskDueDate }

// SetDueDate sets the task's due date.
func (t *Task) SetDueDate(dueDate *time.Time) { t.TaskDueDate = dueDate }

// LabelsString returns the task's label string.
func (t Task) LabelsString() string { return t.TaskLabels }

// SetLabelsString sets the task's labels string.
func (t *Task) SetLabelsString(labels string) { t.TaskLabels = labels }

// LabelsList returns the task's labels as slice of string.
func (t *Task) LabelsList() []string {
	var result []string

	for _, label := range strings.Split(t.TaskLabels, ",") {
		if label != "" {
			result = append(result, strings.TrimSpace(label))
		}
	}
	return result
}

// InProgress returns true if the task is marked as in progress.
func (t Task) InProgress() bool { return t.TaskInProgress }

// SetInProgress sets the task's in progress status.
func (t *Task) SetInProgress(inProgress bool) { t.TaskInProgress = inProgress }

// Completed returns true if the task is marked as done.
func (t Task) Completed() bool { return t.TaskCompleted }

// SetCompleted sets the task's completion status.
func (t *Task) SetCompleted(completed bool) { t.TaskCompleted = completed }

// FilterValue returns a string used for filtering/search, combining title and labels.
func (t Task) FilterValue() string { return fmt.Sprintf("%s %s", t.TaskTitle, t.TaskLabels) }

// CropTaskTitle returns the task's title cropped to fit
// length with a concatenated ellipses.
func (t Task) CropTaskTitle(length int) string {
	if len(t.Title()) > length {
		return t.TaskTitle[:length-len(ellipses)] + ellipses
	}

	return t.TaskTitle
}

// CropTaskLabels returns the task's labels as string.
// Labels are separated by comma + whitespace.
// If the returned string would exceed length
// it is cropped and an ellipses is appended to fit length.
func (t Task) CropTaskLabels(length int) string {
	if len(t.LabelsString()) > length {
		return strings.ReplaceAll(t.TaskLabels[:length-len(ellipses)]+ellipses, ",", ", ")
	}

	return strings.ReplaceAll(t.TaskLabels, ",", ", ")
}

// DueDateToString formats the task's due date as a string using DueDateLayout.
// Returns an empty string if no due date is set.
func (t Task) DueDateToString() string {
	if t.TaskDueDate != nil {
		return t.DueDate().Format(time.DateTime)
	}

	return ""
}

// DaysUntilToString returns a string containing the full days from now until the due date.
// If the date is in the past, it returns a negative value.
// Returns "no due date" if executed on a task with missing due date.
func (t Task) DaysUntilToString() string {
	if t.TaskDueDate != nil {
		now := time.Now()
		dueDate := t.DueDate()

		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		target := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())

		diff := target.Sub(now).Hours() / 24

		return fmt.Sprintf("%d", int(math.Floor(diff)))
	}

	return "no due date"
}

// PriorityValue returns a numeric value for the task's priority.
// Useful for sorting tasks by urgency.
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

// MarshalTask returns a pretty-printed JSON representation of the task.
// Panics if serialization fails.
func (t Task) MarshalTask() []byte {
	json, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

// WriteTaskJSON writes the given task JSON to disk under the project directory,
// using the task's ID as the filename. Returns a Tea message on success or error.
func (t Task) WriteTaskJSON(json []byte, p Project, kind string) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.ID(), t.ID()+".json")

		if err := os.WriteFile(file, json, 0o600); err != nil {
			return WriteTaskJSONErrorMsg{err}
		}

		return WriteTaskJSONDoneMsg{Task: t, Kind: kind}
	}
}

// DeleteTaskFromFS deletes the task's JSON file from the given project directory.
// Returns a Tea message on success or failure.
func (t *Task) DeleteTaskFromFS(p Project) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.ID(), t.ID()+".json")

		err := os.Remove(file)
		if err != nil {
			return TaskDeleteErrorMsg{err}
		}

		return TaskDeleteDoneMsg{}
	}
}

// TaskToMarkdown returns a Markdown-formatted string representing the task,
// including metadata like description, due date, priority, labels, and completion status.
func (t *Task) TaskToMarkdown() string {
	title := fmt.Sprintf("# %s\n\n", t.Title())

	completed := "✅  **Done**: ❌ No\n\n"
	if t.Completed() {
		completed = "✅  **Done**: ✅ Yes\n\n"
	}

	inProgress := ""
	if !t.Completed() {
		inProgress = "🚧  **In Progress**: ❌ No\n\n"
		if t.InProgress() {
			inProgress = "🚧  **In Progress**: ✅ Yes\n\n"
		}
	}
	inProgress += "---\n\n"

	priority := fmt.Sprintf("🎯  **Priority**: %s\n\n", strings.ToUpper(t.Priority()))

	dueDate := ""
	if t.DueDate() != nil {
		dueDate = fmt.Sprintf("📅  **Due At**: %s\n\n", t.DueDate().Format(time.RFC1123))
	}
	dueDate += "---\n\n"

	description := fmt.Sprintf("📝  **Description**\n\n%s\n\n---\n\n", t.Description())

	labels := ""
	if t.LabelsString() != "" {
		labels = "🏷️  **Labels**\n\n"
		labelsSeq := strings.SplitSeq(t.LabelsString(), ",")
		for label := range labelsSeq {
			labels += "- " + label + "\n\n"
		}
		labels += "\n\n---\n\n"
	}

	id := fmt.Sprintf("🆔  **ID**: %s", t.ID())

	return title + completed + inProgress + priority + dueDate + description + labels + id
}
