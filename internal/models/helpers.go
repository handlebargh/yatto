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

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"
)

// completedString returns a string representation of the task completion state.
// It returns "done" if completed is true, otherwise "open".
func completedString(completed bool) string {
	if completed {
		return "done"
	}

	return "open"
}

// getColorCode maps a project color name to its corresponding lipgloss.AdaptiveColor.
// Supported colors include: green, orange, red, blue, indigo.
// Defaults to blue if the color is unrecognized.
func getColorCode(color string) lipgloss.AdaptiveColor {
	switch color {
	case "green":
		return green
	case "orange":
		return orange
	case "red":
		return red
	case "blue":
		return blue
	case "indigo":
		return indigo
	default:
		return blue
	}
}

// taskSortValue returns a numeric value used to sort tasks by priority and completion status.
// Lower values indicate higher priority. Completed tasks are deprioritized by adding a fixed offset.
func taskSortValue(t *items.Task) int {
	base := 10 - t.PriorityValue()
	if t.Completed() {
		base += 100
	}
	return base
}

// tickCmd returns a Bubble Tea command that sends a tickMsg every second.
// Used to drive periodic updates in the TUI, such as progress animations.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// readProjectsFromFS reads all project directories from the configured storage path.
// It deserializes each project's `project.json` file into an items.Project object.
// Returns a slice of all successfully read projects.
// Panics if the storage directory can't be read or if project files are invalid.
func readProjectsFromFS() []items.Project {
	storageDir := viper.GetString("storage.path")
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var projects []items.Project
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}

		dirPath := filepath.Join(storageDir, entry.Name())

		projectFile, err := os.ReadFile(filepath.Join(dirPath, "project.json"))
		if err != nil {
			panic(err)
		}

		var project items.Project
		if err := json.Unmarshal(projectFile, &project); err != nil {
			panic(err)
		}
		projects = append(projects, project)
	}

	return projects
}
