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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

type (
	// WriteProjectJSONDoneMsg indicates successful write of a project JSON file.
	WriteProjectJSONDoneMsg struct {
		Project Project
		Kind    string
	}

	// WriteProjectJSONErrorMsg is returned when a project fails to serialize or write to disk.
	WriteProjectJSONErrorMsg struct{ Err error }

	// ProjectDeleteDoneMsg indicates successful deletion of a project directory.
	ProjectDeleteDoneMsg struct{}

	// ProjectDeleteErrorMsg is returned when a project fails to delete from disk.
	ProjectDeleteErrorMsg struct{ Err error }
)

// Error implements the error interface for WriteProjectJSONErrorMsg.
func (e WriteProjectJSONErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for ProjectDeleteErrorMsg.
func (e ProjectDeleteErrorMsg) Error() string { return e.Err.Error() }

// Project represents a collection of tasks, identified by an ID, title, description,
// and a display color. Projects are stored as directories on disk containing a JSON file
// holding the data defined in the Project type.
type Project struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color"`
}

// FilterValue returns a string used for filtering/search, based on project title.
func (p *Project) FilterValue() string { return p.Title }

// ReadTasksFromFS reads all task files from the project's directory
// and returns them as a slice of Task. It panics if the directory
// or any task file cannot be read or parsed.
func (p *Project) ReadTasksFromFS() []Task {
	storageDir := viper.GetString("storage.path")
	taskFiles, err := os.ReadDir(filepath.Join(storageDir, p.ID))
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		if entry.IsDir() || !UUIDRegex.MatchString(entry.Name()) {
			continue
		}

		filePath := filepath.Join(storageDir, p.ID, entry.Name())
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

// DeleteProjectFromFS deletes the entire project directory and all its contents
// from disk. Returns a Tea message indicating success or failure.
func (p *Project) DeleteProjectFromFS() tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Join(viper.GetString("storage.path"), p.ID)

		err := os.RemoveAll(dir)
		if err != nil {
			return ProjectDeleteErrorMsg{err}
		}

		return ProjectDeleteDoneMsg{}
	}
}

// MarshalProject returns a pretty-printed JSON representation of the project.
// Panics if the project cannot be serialized.
func (p *Project) MarshalProject() []byte {
	bytes, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		panic(err)
	}

	return bytes
}

// WriteProjectJSON writes the given project JSON to disk as project.json
// inside the project's directory. Ensures the directory exists.
// Returns a Tea message indicating success or error.
func (p *Project) WriteProjectJSON(json []byte, kind string) tea.Cmd {
	return func() tea.Msg {
		storageDir := viper.GetString("storage.path")

		// ensure project directory
		dir := filepath.Join(storageDir, p.ID)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		file := filepath.Join(storageDir, p.ID, "project.json")
		if err := os.WriteFile(file, json, 0o600); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		return WriteProjectJSONDoneMsg{Project: *p, Kind: kind}
	}
}

// NumOfTasks returns counts for all tasks in the project directory:
// - total number of tasks
// - number of completed tasks
// - number of tasks due today
//
// Returns an error if the directory cannot be read or if a task cannot be parsed.
func (p *Project) NumOfTasks() (int, int, int, error) {
	dir := filepath.Join(viper.GetString("storage.path"), p.ID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, 0, err
	}

	total, completed, due := 0, 0, 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "project.json" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var t struct {
			DueDate   *time.Time `json:"due_date"`
			Completed bool       `json:"completed"`
		}
		if err := json.Unmarshal(data, &t); err != nil {
			return 0, 0, 0, err
		}

		total++

		if t.Completed {
			completed++
		} else if IsToday(t.DueDate) {
			due++
		}

	}

	return total, completed, due, nil
}
