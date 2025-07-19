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
	ProjectDescription string `json:"description"`
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

func ReadProjectsFromFS() []Project {
	storageDir := viper.GetString("storage.path")
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}

		dirPath := filepath.Join(storageDir, entry.Name())

		projectFile, err := os.ReadFile(filepath.Join(dirPath, "project.json"))
		if err != nil {
			panic(err)
		}

		var project Project
		if err := json.Unmarshal(projectFile, &project); err != nil {
			panic(err)
		}
		projects = append(projects, project)
	}

	return projects
}

func DeleteProjectFromFS(project *Project) tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Join(viper.GetString("storage.path"), project.Id())

		err := os.RemoveAll(dir)
		if err != nil {
			return ProjectDeleteErrorMsg{err}
		}

		return ProjectDeleteDoneMsg{}
	}
}

func MarshalProject(uuid, title, description, color string) []byte {
	var project Project
	project.SetId(uuid)
	project.SetTitle(title)
	project.SetDescription(description)
	project.SetColor(color)

	json, err := json.MarshalIndent(project, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func WriteProjectJson(json []byte, project Project, kind string) tea.Cmd {
	return func() tea.Msg {
		storageDir := viper.GetString("storage.path")

		// ensure project directory
		dir := filepath.Join(storageDir, project.Id())
		if err := os.Mkdir(dir, 0700); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		file := filepath.Join(storageDir, project.Id(), "project.json")
		if err := os.WriteFile(file, json, 0600); err != nil {
			return WriteProjectJSONErrorMsg{err}
		}

		return WriteProjectJSONDoneMsg{Project: project, Kind: kind}
	}
}
