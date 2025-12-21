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
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/spf13/viper"
)

func TestE2E_AddEditDeleteTask(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		projectTitle string
		projectDesc  string
		taskTitle    string
		taskDesc     string
		append       string
		cfg          func(*testing.T) *viper.Viper
	}{
		{"git",
			"Test1",
			"Desc1",
			"Test task 1",
			"Test task 1 description",
			" edited",
			setGitAppConfig},
		{"jj",
			"Test1",
			"Desc1",
			"Test task 1",
			"Test task 1 description",
			" edited",
			setJJAppConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newE2E(t, tc.cfg(t))

			// First create and enter a new project.
			e.addProject(tc.projectTitle, tc.projectDesc, []string{tc.projectTitle, "Projects"})
			e.chooseItem(tc.projectTitle, false)
			e.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

			// Then run the actual test.
			e.addTask(tc.taskTitle, tc.taskDesc, []string{"1 task", tc.taskTitle})
			e.editTask(tc.taskTitle, tc.append, tc.append, []string{"1 task", tc.taskTitle + tc.append})
			e.deleteItems("task", []string{tc.taskTitle + tc.append}, []string{tc.taskTitle + tc.append}, []string{"No tasks"})

			e.tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'q'},
			})
			e.tm.Send(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'q'},
			})

			e.tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
			out := e.tm.FinalModel(t).View()
			teatest.RequireEqualOutput(t, []byte(out))
		})
	}
}
