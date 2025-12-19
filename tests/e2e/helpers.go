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

type e2e struct {
	t  *testing.T
	tm *teatest.TestModel
}

func newE2E(t *testing.T, v *viper.Viper) *e2e {
	t.Helper()

	tm := teatest.NewTestModel(
		t,
		models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(400, 400),
	)

	e := &e2e{t: t, tm: tm}

	e.waitForProjectsScreen()
	return e
}

func (e *e2e) waitForProjectsScreen() {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("No projects"))
	}, teatest.WithDuration(2*time.Second))
}

func (e *e2e) waitForProject(name string) {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("1 project")) &&
			bytes.Contains(bts, []byte(name))
	}, teatest.WithDuration(2*time.Second))
}

func (e *e2e) waitForProjectGone(name string) {
	e.t.Helper()

	teatest.WaitFor(e.t, e.tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("No projects")) &&
			!bytes.Contains(bts, []byte(name))
	}, teatest.WithDuration(2*time.Second))
}

func (e *e2e) addProject(name, desc string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(desc)})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	e.waitForProject(name)
}

func (e *e2e) editProject(name, appendText string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(appendText)})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	e.waitForProject(name + appendText)
}

func (e *e2e) deleteProject(name string) {
	e.t.Helper()

	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/" + name)})
	e.tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	e.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	e.waitForProjectGone(name)
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
