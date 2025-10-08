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
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
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
	TaskDeleteDoneMsg struct{ Task Task }

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
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Priority    string     `json:"priority"`
	Labels      string     `json:"labels,omitempty"`
	Author      string     `json:"author,omitempty"`
	Assignee    string     `json:"assignee,omitempty"`
	InProgress  bool       `json:"in_progress"`
	Completed   bool       `json:"completed"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// LabelsList returns the task's labels as slice of string.
func (t *Task) LabelsList() []string {
	var result []string

	for _, label := range strings.Split(t.Labels, ",") {
		if label != "" {
			result = append(result, strings.TrimSpace(label))
		}
	}
	return result
}

// FilterValue returns a string used for filtering/search, combining title and labels.
func (t *Task) FilterValue() string { return fmt.Sprintf("%s %s", t.Title, t.Labels) }

// CropTaskTitle returns the task's title cropped to fit
// length with a concatenated ellipses.
func (t *Task) CropTaskTitle(length int) string {
	if len(t.Title) > length {
		return t.Title[:length-len(ellipses)] + ellipses
	}

	return t.Title
}

// CropTaskLabels returns the task's labels as string.
// Labels are separated by comma + whitespace.
// If the returned string would exceed length
// it is cropped and an ellipses is appended to fit length.
func (t *Task) CropTaskLabels(length int) string {
	if len(t.Labels) > length {
		return strings.ReplaceAll(t.Labels[:length-len(ellipses)]+ellipses, ",", ", ")
	}

	labels := strings.ReplaceAll(t.Labels, ",", ", ")

	if labels == "" {
		return "No labels"
	}

	return labels
}

// DueDateToString formats the task's due date as a string using DueDateLayout.
// Returns an empty string if no due date is set.
func (t *Task) DueDateToString() string {
	if t.DueDate != nil {
		return t.DueDate.Format(time.DateTime)
	}

	return ""
}

// DaysUntilToString returns a string containing the full days from now until the due date.
// If the date is in the past, it returns a negative value.
// Returns "no due date" if executed on a task with missing due date.
func (t *Task) DaysUntilToString() string {
	if t.DueDate != nil {
		now := time.Now()
		dueDate := t.DueDate

		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		target := time.Date(
			dueDate.Year(),
			dueDate.Month(),
			dueDate.Day(),
			0,
			0,
			0,
			0,
			dueDate.Location(),
		)

		diff := target.Sub(now).Hours() / 24

		return fmt.Sprintf("%d", int(math.Floor(diff)))
	}

	return "no due date"
}

// PriorityValue returns a numeric value for the task's priority.
// Useful for sorting tasks by urgency.
func (t *Task) PriorityValue() int {
	switch t.Priority {
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
func (t *Task) MarshalTask() []byte {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "\t")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(t); err != nil {
		panic(err)
	}

	// Remove the trailing newline added by Encode
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
}

// WriteTaskJSON writes the given task JSON to disk under the project directory,
// using the task's ID as the filename. Returns a Tea message on success or error.
func (t *Task) WriteTaskJSON(json []byte, p Project, kind string) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.ID, t.ID+".json")

		if err := os.WriteFile(file, json, 0o600); err != nil {
			return WriteTaskJSONErrorMsg{err}
		}

		return WriteTaskJSONDoneMsg{Task: *t, Kind: kind}
	}
}

// DeleteTaskFromFS deletes the task's JSON file from the given project directory.
// Returns a Tea message on success or failure.
func (t *Task) DeleteTaskFromFS(p Project) tea.Cmd {
	return func() tea.Msg {
		file := filepath.Join(viper.GetString("storage.path"), p.ID, t.ID+".json")

		err := os.Remove(file)
		if err != nil {
			return TaskDeleteErrorMsg{err}
		}

		return TaskDeleteDoneMsg{*t}
	}
}

// FindListIndexByID returns the index of the task in the given slice of list.Item,
// or -1 if not found.
func (t *Task) FindListIndexByID(items []list.Item) int {
	for i, item := range items {
		task, ok := item.(*Task)
		if ok && task.ID == t.ID {
			return i
		}
	}

	return -1 // not found
}

// TaskToMarkdown returns a Markdown-formatted string representing the task,
// including metadata like description, due date, priority, labels, and completion status.
func (t *Task) TaskToMarkdown() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", t.Title))
	content.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", t.Description))

	content.WriteString("## Metadata\n\n")
	content.WriteString("| **Completed** | **In Progress** | **Priority** |\n")
	content.WriteString("| ------------- | --------------- | ------------ |\n")

	completed := "NO"
	if t.Completed {
		completed = "YES"
	}

	inProgress := "NO"
	if !t.Completed && t.InProgress {
		inProgress = "YES"
	}

	priority := strings.ToUpper(t.Priority)

	content.WriteString("|" + completed + "|" + inProgress + "|" + priority + "\n\n")

	if t.Author != "" {
		content.WriteString("| **Task Author** |\n")
		content.WriteString("| --------------- |\n\n")
		content.WriteString(t.Author)
		content.WriteString("\n\n")
	}

	if t.Assignee != "" {
		content.WriteString("| **Assigned to** |\n")
		content.WriteString("| --------------- |\n\n")
		content.WriteString(t.Assignee)
		content.WriteString("\n\n")
	}

	if t.DueDate != nil {
		content.WriteString("| **Due Date** |\n")
		content.WriteString("| ------------ |\n")
		content.WriteString(fmt.Sprintf("| %s |\n\n", t.DueDate.Format(time.RFC1123)))
	}

	if t.Labels != "" {
		content.WriteString("| **Labels** |\n")
		content.WriteString("| ---------- |\n")

		labelsSeq := strings.SplitSeq(t.Labels, ",")
		for label := range labelsSeq {
			content.WriteString(fmt.Sprintf("| - %s |\n", label))
		}
		content.WriteString("\n")
	}

	content.WriteString("| **ID** |\n")
	content.WriteString("| ------ |\n")
	content.WriteString(fmt.Sprintf("| %s |\n\n", t.ID))

	return content.String()
}
