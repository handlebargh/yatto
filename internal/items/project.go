package items

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// NumOfTasksInProject calculates the total of tasks
// for a given project.
func (p Project) NumTotalTasks() (int, error) {
	entries, err := os.ReadDir(filepath.Join(
		viper.GetString("storage.path"), p.Id()))
	if err != nil {
		return 0, err
	}

	tasks := 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "project.json" {
			continue
		}

		tasks++
	}

	return tasks, nil
}

// NumOfCompletedTasksInProject calculates the number of tasks
// not yet completed for a given project.
func (p Project) NumCompletedTasks() (int, error) {
	dir := filepath.Join(viper.GetString("storage.path"), p.Id())
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	completedTasks := 0
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
			Completed bool `json:"completed"`
		}
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}

		if t.Completed {
			completedTasks++
		}
	}

	return completedTasks, nil
}
