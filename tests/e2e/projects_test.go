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

package e2e

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/handlebargh/yatto/internal/models"
)

func TestE2E_AddAndEditProjectGit(t *testing.T) {
	t.Parallel()

	v := setGitAppConfig(t)

	tm := teatest.NewTestModel(t, models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(400, 400),
	)

	// Add a project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Test Project")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("This is a test project.")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	// Edit the project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" edited")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t)
	finalOutput := tm.FinalOutput(t)

	finalBytes, err := io.ReadAll(finalOutput)
	if err != nil {
		t.Fatalf("failed to get final view: %v", err)
	}

	if !strings.Contains(string(finalBytes), "Test Project edited") {
		t.Errorf("expected to find 'Test Project edited' in the final view, but didn't")
	}
}

func TestE2E_AddAndEditProjectJJ(t *testing.T) {
	t.Parallel()

	v := setJJAppConfig(t)

	tm := teatest.NewTestModel(t, models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(400, 400),
	)

	// Add a project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Test Project")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("This is a test project.")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	// Edit the project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" edited")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t)
	finalOutput := tm.FinalOutput(t)

	finalBytes, err := io.ReadAll(finalOutput)
	if err != nil {
		t.Fatalf("failed to get final view: %v", err)
	}

	if !strings.Contains(string(finalBytes), "Test Project edited") {
		t.Errorf("expected to find 'Test Project edited' in the final view, but didn't")
	}
}

func TestE2E_AddAndDeleteProjectGit(t *testing.T) {
	t.Parallel()

	v := setGitAppConfig(t)

	tm := teatest.NewTestModel(t, models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(400, 400),
	)

	// Add a project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Test Project")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("This is a test project.")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	// Delete the project
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t)
	finalOutput := tm.FinalOutput(t)

	finalBytes, err := io.ReadAll(finalOutput)
	if err != nil {
		t.Fatalf("failed to get final view: %v", err)
	}

	if !strings.Contains(string(finalBytes), "No projects") {
		t.Errorf("expected to find 'No projects' in the final view, but didn't")
	}
}

func TestE2E_AddAndDeleteProjectJJ(t *testing.T) {
	t.Parallel()

	v := setJJAppConfig(t)

	tm := teatest.NewTestModel(t, models.InitialProjectListModel(v),
		teatest.WithInitialTermSize(400, 400),
	)

	// Add a project
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Test Project")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("This is a test project.")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	// Delete the project
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	time.Sleep(2 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t)
	finalOutput := tm.FinalOutput(t)

	finalBytes, err := io.ReadAll(finalOutput)
	if err != nil {
		t.Fatalf("failed to get final view: %v", err)
	}

	if !strings.Contains(string(finalBytes), "No projects") {
		t.Errorf("expected to find 'No projects' in the final view, but didn't")
	}
}
