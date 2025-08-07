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

// package helpers defines some functions
// used by several other packages.
package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"
)

// ReadProjectsFromFS reads all project directories from the configured storage path.
// It deserializes each project's `project.json` file into an items.Project object.
// Returns a slice of all successfully read projects.
// Panics if the storage directory can't be read or if project files are invalid.
func ReadProjectsFromFS() []items.Project {
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

// GetColorCode maps a project color name to its corresponding lipgloss.AdaptiveColor.
// Supported colors include: green, orange, red, blue, indigo.
// Defaults to blue if the color is unrecognized.
func GetColorCode(color string) lipgloss.AdaptiveColor {
	switch color {
	case "green":
		return colors.Green()
	case "orange":
		return colors.Orange()
	case "red":
		return colors.Red()
	case "blue":
		return colors.Blue()
	case "indigo":
		return colors.Indigo()
	default:
		return colors.Blue()
	}
}
