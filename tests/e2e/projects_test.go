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

func TestE2E_AddEditDeleteProject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  func(*testing.T) *viper.Viper
	}{
		{"git", setGitAppConfig},
		{"jj", setJJAppConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newE2E(t, tc.cfg(t))

			e.addProject("Test Project", "This is a test project", []string{"Projects", "Test Project"})
			e.editProject("Test Project", " edited", []string{"Projects", "Test Project edited"})
			e.deleteProject("Test Project edited", []string{"Test Project edited"}, []string{"No projects"})

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
