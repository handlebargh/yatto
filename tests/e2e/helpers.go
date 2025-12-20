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

// Package e2e contains all end to end tests.
package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/handlebargh/yatto/internal/models"
	"github.com/spf13/viper"
)

const defaultWait = 2 * time.Second

type e2e struct {
	t  *testing.T
	tm *teatest.TestModel
}

func newE2E(t *testing.T, v *viper.Viper) *e2e {
	t.Helper()

	tm := teatest.NewTestModel(
		t,
		models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(300, 100),
	)

	e := &e2e{t: t, tm: tm}

	e.waitForProjectsScreen()
	return e
}

func (e *e2e) waitForProjectsScreen() {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("No projects"))
	}, teatest.WithDuration(defaultWait))
}

// waitForMessagesPresent waits until all `present` messages appear in output.
// Empty slices impose no constraint.
func (e *e2e) waitForMessagesPresent(present []string) {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		for _, msg := range present {
			if !bytes.Contains(bts, []byte(msg)) {
				return false
			}
		}

		return true
	}, teatest.WithDuration(defaultWait))
}

// waitForMessageGone waits until all `present` messages appear in output
// and none of the `gone` messages appear. Empty slices impose no constraint.
func (e *e2e) waitForMessageGone(gone, present []string) {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		for _, msg := range present {
			if !bytes.Contains(bts, []byte(msg)) {
				return false
			}
		}

		for _, msg := range gone {
			if bytes.Contains(bts, []byte(msg)) {
				return false
			}
		}

		return true
	}, teatest.WithDuration(defaultWait))
}

func (e *e2e) confirmField(label, value string) {
	e.waitForMessagesPresent([]string{label})
	if value != "" {
		e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)})
	}
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
}

func (e *e2e) enterProject(title string) {
	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/" + title)})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
}

func (e *e2e) addProject(title, desc string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	e.confirmField("Select a color", "")
	e.confirmField("Enter a title", title)
	e.confirmField("Enter a description", desc)
	e.confirmField("Create new project?", "y")

	e.waitForMessagesPresent([]string{"Projects", title})
}

func (e *e2e) editProject(title, appendText string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	e.confirmField("Select a color", "")
	e.confirmField("Enter a title", appendText)
	e.confirmField("Enter a description", "")
	e.confirmField("Edit project?", "y")

	e.waitForMessagesPresent([]string{"Projects", title + appendText})
}

func (e *e2e) deleteProject(title string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	e.confirmField("Filter", title)
	e.waitForMessagesPresent([]string{title})
	e.tm.Send(tea.KeyMsg{Type: tea.KeySpace})

	e.waitForMessagesPresent([]string{"⟹"})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	e.waitForMessagesPresent([]string{"Delete 1 project"})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	e.waitForMessageGone([]string{title}, []string{"No projects"})
}

func (e *e2e) addTask(title, desc string) {
	e.t.Helper()

	e.addProject("Test1", "Desc1")
	e.enterProject("Test1")

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) // Open task creation form

	e.confirmField("Select priority", "")
	e.confirmField("Enter a title", title)
	e.confirmField("Enter a description", desc)
	e.confirmField("Due Date", "")
	e.confirmField("Choose labels", "")
	e.confirmField("Enter additional labels", "")
	e.confirmField("Enter the task author", "")
	e.confirmField("Choose an assignee", "")
	e.confirmField("Enter a new email address", "")
	e.confirmField("Create task?", "y")

	e.waitForMessagesPresent([]string{"1 task", title})
}

func (e *e2e) editTask(appendTitle, appendDesc string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}) // Open task editing form

	e.confirmField("Select priority", "")
	e.confirmField("Enter a title", appendTitle)
	e.confirmField("Enter a description", appendDesc)
	e.confirmField("Due Date", "")
	e.confirmField("Choose labels", "")
	e.confirmField("Enter additional labels", "")
	e.confirmField("Enter the task author", "")
	e.confirmField("Choose an assignee", "")
	e.confirmField("Enter a new email address", "")
	e.confirmField("Edit task?", "y")

	e.waitForMessagesPresent([]string{"1 task", appendTitle})
}

// setGitAppConfig initializes a fresh git repo for testing and sets the viper
// config accordingly.
// Return the path to the testing storage directory.
func setGitAppConfig(t *testing.T) *viper.Viper {
	t.Helper()
	storagePath := setupGitRepo(t)
	v := viper.New()

	v.Set("storage.path", storagePath)
	v.Set("vcs.backend", "git")

	return v
}

// setJJAppConfig initializes a fresh jj repo for testing and sets the viper
// config accordingly.
// Return the path to the testing storage directory.
func setJJAppConfig(t *testing.T) *viper.Viper {
	t.Helper()
	storagePath := setupJJRepo(t)
	v := viper.New()

	v.Set("storage.path", storagePath)
	v.Set("vcs.backend", "jj")

	return v
}

// setupGitRepo creates a temporary directory and initializes a fresh git repo.
// It returns the path to the repo and ensures local git configs don't interfere.
func setupGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	runCmd(t, tmpDir, "git", "init", "--initial-branch", "main")
	runCmd(t, tmpDir, "git", "config", "user.name", "Test User")
	runCmd(t, tmpDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, tmpDir, "git", "config", "commit.gpgSign", "false")

	testFile := filepath.Join(tmpDir, "INIT")
	if err := os.WriteFile(testFile, []byte(""), 0o600); err != nil {
		t.Fatal("error writing INIT file")
	}

	runCmd(t, tmpDir, "git", "add", "INIT")
	runCmd(t, tmpDir, "git", "commit", "-m", "Initial commit")

	return tmpDir
}

// setupJJRepo creates a temporary directory and initializes a fresh jj repo.
// It returns the path to the repo and ensures local jj configs don't interfere.
func setupJJRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	runCmd(t, tmpDir, "jj", "git", "init")
	runCmd(t, tmpDir, "jj", "config", "set", "--repo", "user.name", "Test User")
	runCmd(t, tmpDir, "jj", "config", "set", "--repo", "user.email", "test@example.com")

	testFile := filepath.Join(tmpDir, "INIT")
	if err := os.WriteFile(testFile, []byte(""), 0o600); err != nil {
		t.Fatal("error writing INIT file")
	}

	runCmd(t, tmpDir, "jj", "commit", "--message", "Initial commit")

	return tmpDir
}

// runCmd is a helper to run commands inside the temp directory.
func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run %s %v: %v", name, args, err)
	}
}
