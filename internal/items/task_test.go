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

package items

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func TestTask_CropTaskTitle(t *testing.T) {
	task := &Task{Title: "This is a long title"}
	cropped := task.CropTaskTitle(10)
	if !strings.HasSuffix(cropped, "...") {
		t.Errorf("Expected title to be cropped with an ellipsis, but got %s", cropped)
	}
}

func TestTask_CropTaskLabels(t *testing.T) {
	task := &Task{Labels: "label1,label2,label3"}
	cropped := task.CropTaskLabels(10)
	if !strings.HasSuffix(cropped, "...") {
		t.Errorf("Expected labels to be cropped with an ellipsis, but got %s", cropped)
	}
}

func TestTask_DueDateToString(t *testing.T) {
	now := time.Now()
	task := &Task{DueDate: &now}
	expected := now.Format(time.DateTime)
	if task.DueDateToString() != expected {
		t.Errorf("Expected due date to be %s, but got %s", expected, task.DueDateToString())
	}
}

func TestTask_DaysUntilToString(t *testing.T) {
	now := time.Now()
	task := &Task{DueDate: &now}
	if task.DaysUntilToString() != "0" {
		t.Errorf("Expected 0 days until due date, but got %s", task.DaysUntilToString())
	}

	tomorrow := now.AddDate(0, 0, 1)
	task.DueDate = &tomorrow
	if task.DaysUntilToString() != "1" {
		t.Errorf("Expected 1 day until due date, but got %s", task.DaysUntilToString())
	}
}

func TestTask_PriorityValue(t *testing.T) {
	task := &Task{Priority: "high"}
	if task.PriorityValue() != 2 {
		t.Errorf("Expected priority value to be 2, but got %d", task.PriorityValue())
	}

	task.Priority = "medium"
	if task.PriorityValue() != 1 {
		t.Errorf("Expected priority value to be 1, but got %d", task.PriorityValue())
	}

	task.Priority = "low"
	if task.PriorityValue() != 0 {
		t.Errorf("Expected priority value to be 0, but got %d", task.PriorityValue())
	}

	task.Priority = "unknown"
	if task.PriorityValue() != -1 {
		t.Errorf("Expected priority value to be -1, but got %d", task.PriorityValue())
	}
}

func TestTask_TaskToMarkdown(t *testing.T) {
	dueDate := time.Now()

	task := &Task{
		Title:       "Test Task",
		Description: "This is a test task.",
		Priority:    "high",
		Labels:      "label1,label2",
		Author:      "Test User <test.user@example.com>",
		Assignee:    "Test User <test.user@example.com>",
		InProgress:  true,
		Completed:   false,
		DueDate:     &dueDate,
	}
	markdown := task.TaskToMarkdown()
	if !strings.Contains(markdown, "# Test Task") {
		t.Errorf("Expected markdown to contain the task title, but it didn't")
	}
	if !strings.Contains(markdown, "This is a test task.") {
		t.Errorf("Expected markdown to contain the task description, but it didn't")
	}
	if !strings.Contains(markdown, "|NO|YES|HIGH") {
		t.Errorf("Expected markdown to contain the task metadata, but it didn't")
	}
}

func TestTask_WriteTaskJSON(t *testing.T) {
	tempDir := t.TempDir()
	v := viper.New()
	v.Set("storage.path", tempDir)

	project := Project{ID: "test-project"}
	projectDir := filepath.Join(tempDir, project.ID)
	_ = os.Mkdir(projectDir, 0o755)

	task := &Task{ID: uuid.NewString(), Title: "Test Task"}
	cmd := task.WriteTaskJSON(v, task.MarshalTask(), project, "create")
	msg := cmd()

	if _, ok := msg.(WriteTaskJSONDoneMsg); !ok {
		t.Errorf("Expected WriteTaskJSONDoneMsg, but got %T", msg)
	}

	taskFile := filepath.Join(projectDir, task.ID+".json")
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		t.Errorf("Expected task file to be created, but it wasn't")
	}
}

func TestTask_DeleteTaskFromFS(t *testing.T) {
	tempDir := t.TempDir()
	v := viper.New()
	v.Set("storage.path", tempDir)

	project := Project{ID: "test-project"}
	projectDir := filepath.Join(tempDir, project.ID)
	_ = os.Mkdir(projectDir, 0o755)

	task := &Task{ID: uuid.NewString(), Title: "Test Task"}
	taskFile := filepath.Join(projectDir, task.ID+".json")
	_ = os.WriteFile(taskFile, task.MarshalTask(), 0o644)

	cmd := task.DeleteTaskFromFS(v, project)
	msg := cmd()

	if _, ok := msg.(TaskDeleteDoneMsg); !ok {
		t.Errorf("Expected TaskDeleteDoneMsg, but got %T", msg)
	}

	if _, err := os.Stat(taskFile); !os.IsNotExist(err) {
		t.Errorf("Expected task file to be deleted, but it wasn't")
	}
}
