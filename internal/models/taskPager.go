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
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// taskPagerModel represents the Bubble Tea model for the task detail view.
type taskPagerModel struct {
	listModel *taskListModel
	content   string
	ready     bool
	viewport  viewport.Model
}

// newTaskPagerModel creates a new taskPagerModel for the given task content.
func newTaskPagerModel(content string, listModel *taskListModel) taskPagerModel {
	return taskPagerModel{
		listModel: listModel,
		content:   content,
		ready:     false,
	}
}

// Init initializes the taskPagerModel and returns an initial command.
func (m taskPagerModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the taskPagerModel accordingly.
func (m taskPagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if k := msg.String(); k == "q" || k == "esc" || k == "h" {
			return m.listModel, nil
		}

	case tea.WindowSizeMsg:
		footerHeight := lipgloss.Height(m.footerView())

		if !m.ready {
			rendered, err := m.listModel.projectModel.renderer.Render(m.content)
			if err != nil {
				rendered = "Error rendering markdown"
			}

			m.viewport = viewport.New(msg.Width, msg.Height-footerHeight)
			m.viewport.YPosition = 10
			m.viewport.SetContent(rendered)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - footerHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns the string representation of the task detail view.
func (m taskPagerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s", m.viewport.View(), m.footerView())
}

// footerView returns the string representation of the task detail view's footer.
func (m taskPagerModel) footerView() string {
	info := lipgloss.NewStyle().Padding(0, 1).Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
