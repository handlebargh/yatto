// Copyright 2025 handlebargh
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
	WriteProjectJSONDoneMsg struct {
		Project Project
		Kind    string
	}
	WriteProjectJSONErrorMsg struct{ Err error }
	ProjectDeleteDoneMsg     struct{}
	ProjectDeleteErrorMsg    struct{ Err error }
)

func (e WriteProjectJSONErrorMsg) Error() string { return e.Err.Error() }
func (e ProjectDeleteErrorMsg) Error() string    { return e.Err.Error() }

type Project struct {
	ProjectId          string `json:"id"`
	ProjectTitle       string `json:"title"`
	ProjectDescription string `json:"description,omitempty"`
	ProjectColor       string `json:"color"`
}

func (p Project) Id() string                         { return p.ProjectId }
func (p *Project) SetId(id string)                   { p.ProjectId = id }
func (p Project) Title() string                      { return p.ProjectTitle }
func (p *Project) SetTitle(title string)             { p.ProjectTitle = title }
func (p Project) Description() string                { return p.ProjectDescription }
func (p *Project) SetDescription(description string) { p.ProjectDescription = description }
func (p Project) Color() string                      { return p.ProjectColor }
func (p *Project) SetColor(color string)             { p.ProjectColor = color }
func (p Project) FilterValue() string                { return p.ProjectTitle }

// ReadTasksFromFS reads all tasks from the storage directory
// and returns them as a slice of Task.
func (p *Project) ReadTasksFromFS() []Task {
	storageDir := viper.GetString("storage.path")
	taskFiles, err := os.ReadDir(filepath.Join(storageDir, p.Id()))
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		if entry.IsDir() || !uuidRegex.MatchString(entry.Name()) {
			continue
		}

		filePath := filepath.Join(storageDir, p.Id(), entry.Name())
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

func (p *Project) DeleteProjectFromFS() tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Join(viper.GetString("storage.path"), p.Id())

		err := os.RemoveAll(dir)
		if err != nil {
			return ProjectDeleteErrorMsg{err}
		}

		return ProjectDeleteDoneMsg{}
	}
}

func (p Project) MarshalProject() []byte {
	json, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func (p Project) WriteProjectJson(json []byte, kind string) tea.Cmd {
	return func() tea.Msg {
		storageDir := viper.GetString("storage.path")

		// ensure project directory
		dir := filepath.Join(storageDir, p.Id())
		if err := os.MkdirAll(dir, 0700); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		file := filepath.Join(storageDir, p.Id(), "project.json")
		if err := os.WriteFile(file, json, 0600); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		return WriteProjectJSONDoneMsg{Project: p, Kind: kind}
	}
}

// NumOfTasksInProject calculates the total number of tasks,
// the number of completed tasks and the number of tasks due today
// for a given project.
func (p Project) NumOfTasks() (int, int, int, error) {
	dir := filepath.Join(viper.GetString("storage.path"), p.Id())
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
		} else {
			if IsToday(t.DueDate) {
				due++
			}
		}
	}

	return total, completed, due, nil
}
