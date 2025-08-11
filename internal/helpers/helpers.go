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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

// GetAllLabels walks the task storage directory (as configured by the
// "storage.path" setting in Viper), reads all task JSON files whose
// filenames match the UUID pattern, and extracts their labels.
//
// Each label found is counted, and the function returns a map where the
// keys are label strings and the values are the number of times each
// label appears across all tasks.
//
// Task files are expected to contain a "labels" field as a comma-separated
// string. This string is split and trimmed by labelsStringToSlice before
// counting.
//
// Errors encountered while reading or parsing individual files are logged
// to standard error, but do not stop processing of other files.
func AllLabels() map[string]int {
	storageDir := viper.GetString("storage.path")

	// Store labels in a map and track their occurrence.
	labelCount := make(map[string]int)

	err := filepath.WalkDir(storageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !items.UUIDRegex.MatchString(filepath.Base(path)) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "error reading task labels at file %s: %v\n", path, readErr)
			return nil
		}

		var task struct {
			Labels string `json:"labels"`
		}

		if parseErr := json.Unmarshal(data, &task); parseErr != nil {
			fmt.Fprintf(os.Stderr, "error parsing json at file %s: %v\n", path, parseErr)
			return nil
		}

		for _, label := range LabelsStringToSlice(task.Labels) {
			labelCount[label]++
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task labels: %v\n", err)
	}

	return labelCount
}

// labelsStringToSlice splits a comma-separated labels string into a slice of
// individual labels. Each label in the result is trimmed of leading and trailing
// whitespace. Empty entries are discarded.
//
// Example:
//
//	input := "work, urgent ,home "
//	output := labelsStringToSlice(input)
//	// output: []string{"work", "urgent", "home"}
func LabelsStringToSlice(labels string) []string {
	var result []string

	for _, label := range strings.Split(labels, ",") {
		if label != "" {
			result = append(result, strings.TrimSpace(label))
		}
	}

	return result
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
