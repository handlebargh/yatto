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

func TestProject_CropDescription(t *testing.T) {
	project := &Project{Description: "This is a long description"}
	cropped := project.CropDescription(10)
	if !strings.HasSuffix(cropped, "...") {
		t.Errorf("Expected description to be cropped with an ellipsis, but got %s", cropped)
	}
}

func TestProject_WriteProjectJSON(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	project := &Project{ID: "test-project", Title: "Test Project"}
	cmd := project.WriteProjectJSON(project.MarshalProject(), "create")
	msg := cmd()

	if _, ok := msg.(WriteProjectJSONDoneMsg); !ok {
		t.Errorf("Expected WriteProjectJSONDoneMsg, but got %T", msg)
	}

	projectDir := filepath.Join(tempDir, project.ID)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Errorf("Expected project directory to be created, but it wasn't")
	}

	jsonFile := filepath.Join(projectDir, "project.json")
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Errorf("Expected project.json to be created, but it wasn't")
	}
}

func TestProject_DeleteProjectFromFS(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	project := &Project{ID: "test-project", Title: "Test Project"}
	projectDir := filepath.Join(tempDir, project.ID)
	_ = os.Mkdir(projectDir, 0o755)

	cmd := project.DeleteProjectFromFS()
	msg := cmd()

	if _, ok := msg.(ProjectDeleteDoneMsg); !ok {
		t.Errorf("Expected ProjectDeleteDoneMsg, but got %T", msg)
	}

	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Errorf("Expected project directory to be deleted, but it wasn't")
	}
}

func TestProject_ReadTasksFromFS(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	project := &Project{ID: "test-project", Title: "Test Project"}
	projectDir := filepath.Join(tempDir, project.ID)
	_ = os.Mkdir(projectDir, 0o755)

	task1 := &Task{ID: uuid.NewString(), Title: "Task 1"}
	task2 := &Task{ID: uuid.NewString(), Title: "Task 2"}

	_ = os.WriteFile(filepath.Join(projectDir, task1.ID+".json"), task1.MarshalTask(), 0o644)
	_ = os.WriteFile(filepath.Join(projectDir, task2.ID+".json"), task2.MarshalTask(), 0o644)

	tasks := project.ReadTasksFromFS()

	if len(tasks) != 2 {
		t.Errorf("Expected to read 2 tasks, but got %d", len(tasks))
	}
}

func TestProject_NumOfTasks(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	project := &Project{ID: "test-project", Title: "Test Project"}
	projectDir := filepath.Join(tempDir, project.ID)
	_ = os.Mkdir(projectDir, 0o755)

	now := time.Now()
	task1 := &Task{ID: uuid.NewString(), Title: "Task 1", Completed: true}
	task2 := &Task{ID: uuid.NewString(), Title: "Task 2", DueDate: &now}
	task3 := &Task{ID: uuid.NewString(), Title: "Task 3"}

	_ = os.WriteFile(filepath.Join(projectDir, task1.ID+".json"), task1.MarshalTask(), 0o644)
	_ = os.WriteFile(filepath.Join(projectDir, task2.ID+".json"), task2.MarshalTask(), 0o644)
	_ = os.WriteFile(filepath.Join(projectDir, task3.ID+".json"), task3.MarshalTask(), 0o644)

	total, completed, due, err := project.NumOfTasks()
	if err != nil {
		t.Fatalf("NumOfTasks returned an error: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total tasks to be 3, but got %d", total)
	}
	if completed != 1 {
		t.Errorf("Expected completed tasks to be 1, but got %d", completed)
	}
	if due != 1 {
		t.Errorf("Expected due tasks to be 1, but got %d", due)
	}
}
