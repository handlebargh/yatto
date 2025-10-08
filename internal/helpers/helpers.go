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

// Package helpers defines some functions
// used by several other packages.
package helpers

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"
)

// ErrUnexpectedInput is returned when a user's input is unexpected.
var ErrUnexpectedInput = errors.New("unexpected input")

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
		if !entry.IsDir() || entry.Name() == ".git" || entry.Name() == ".jj" {
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

// AllLabels walks the task storage directory (as configured by the
// "storage.path" setting in Viper), reads all task JSON files whose
// filenames match the UUID pattern, and extracts their labels.
//
// Each label found is counted, and the function returns a map where the
// keys are label strings and the values are the number of times each
// label appears across all tasks.
//
// Task files are expected to contain a "labels" field as a comma-separated
// string. This string is split and trimmed by LabelsStringToSlice before
// counting.
//
// It is assumed that all matching files are readable and contain valid JSON.
// If this invariant is violated (e.g., a file is unreadable or cannot be
// parsed), the function will panic immediately rather than attempting to
// recover.
func AllLabels() map[string]int {
	storageDir := viper.GetString("storage.path")

	// Store labels in a map and track their frequency.
	labelCount := make(map[string]int)

	err := filepath.WalkDir(storageDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			panic(fmt.Sprintf("unexpected FS walk error at %s: %v", path, walkErr))
		}

		if d.IsDir() || !items.UUIDRegex.MatchString(filepath.Base(path)) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("unexpected read error for %s: %v", path, err))
		}

		var task struct {
			Labels string `json:"labels"`
		}
		if err := json.Unmarshal(data, &task); err != nil {
			panic(fmt.Sprintf("unexpected JSON parse error for %s: %v", path, err))
		}

		for _, label := range LabelsStringToSlice(task.Labels) {
			labelCount[label]++
		}

		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("unexpected error walking storage dir %s: %v", storageDir, err))
	}

	return labelCount
}

// LabelsStringToSlice splits a comma-separated labels string into a slice of
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

// PromptUser writes the given message to the provided output stream and reads a
// single line of input from the provided input stream. The input is trimmed of
// surrounding whitespace and transformed to lower case before being validated.
//
// If no expectedInput values are provided, the trimmed input is returned as-is.
// If one or more expectedInput values are provided, the function only accepts
// input that matches one of them. If the input does not match any allowed
// value, ErrUnexpectedInput is returned.
//
// If writing to the output stream or reading from the input stream fails, an
// error is returned.
//
// Parameters:
//   - input: the reader from which user input is read.
//   - output: the writer to which the prompt message and error messages are written.
//   - message: the prompt message displayed to the user.
//   - expectedInput: an optional list of allowed input values.
//
// Returns the user’s validated input string and an error, if any occurred.
func PromptUser(input io.Reader, output io.Writer, message string, expectedInput ...string) (string, error) {
	reader := bufio.NewReader(input)

	_, err := fmt.Fprint(output, message)
	if err != nil {
		return "", err
	}

	userInput, err := reader.ReadString('\n')
	if err != nil {
		_, err := fmt.Fprintln(output, "An error occurred while reading input. Please try again", err)
		if err != nil {
			return "", err
		}

		return "", err
	}

	userInput = strings.ToLower(strings.TrimSpace(userInput))

	if len(expectedInput) == 0 {
		return userInput, nil
	}

	for _, allowed := range expectedInput {
		userInput = strings.TrimSpace(userInput)
		if userInput == allowed {
			return userInput, nil
		}
	}

	return "", ErrUnexpectedInput
}

// UniqueNonEmptyStrings splits the input string by newlines, trims whitespace from each line,
// and returns a slice of unique, non-empty strings in arbitrary order.
func UniqueNonEmptyStrings(slice []string) []string {
	uniqueItems := make(map[string]bool)

	for _, item := range slice {
		item = strings.TrimSpace(item)
		if item != "" {
			uniqueItems[item] = true
		}
	}

	result := make([]string, 0, len(uniqueItems))
	for item := range uniqueItems {
		result = append(result, AddAngleBracketsToEmail(item))
	}

	return result
}

// AddAngleBracketsToEmail wraps the email address in a string with "<" and ">"
// if it is not already wrapped. The function assumes the email is the last
// word in the string (or the only word containing "@").
//
// If the email is already wrapped in "<" and ">", the string is returned unchanged.
// If no email is found, the string is returned as-is.
func AddAngleBracketsToEmail(s string) string {
	atIndex := strings.Index(s, "@")
	if atIndex == -1 {
		return s
	}

	start := strings.LastIndex(s[:atIndex], " ") + 1
	end := len(s)

	email := s[start:end]

	// Check if the email is already wrapped in < and >
	if strings.HasPrefix(email, "<") && strings.HasSuffix(email, ">") {
		return s
	}

	return s[:start] + "<" + email + ">"
}
