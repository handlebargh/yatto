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

package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/items"
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

		switch {
		case key.Matches(msg, m.listModel.keys.quit) || key.Matches(msg, m.listModel.keys.goBackVim):
			return m.listModel, nil

		case key.Matches(msg, m.listModel.keys.editItem):
			if m.listModel.list.SelectedItem() != nil {
				// Switch to formModel for editing.
				formModel := newTaskFormModel(m.listModel.list.SelectedItem().(*items.Task), m.listModel, true)
				return formModel, tea.WindowSize()
			}

			return m, nil

		case key.Matches(msg, m.listModel.keys.toggleInProgress):
			return m.toggleSelectedTask(
				func(t *items.Task) { t.InProgress = !t.InProgress },
				func(t *items.Task) (bool, string) {
					if t.Completed {
						return false, "Cannot set completed task as in progress"
					}
					return true, ""
				},
				func(t *items.Task) string {
					if t.InProgress {
						return "start"
					}
					return "stop"
				},
				"progress",
			)

		case key.Matches(msg, m.listModel.keys.toggleComplete):
			return m.toggleSelectedTask(
				func(t *items.Task) { t.Completed = !t.Completed; t.InProgress = false },
				func(_ *items.Task) (bool, string) { return true, "" },
				func(t *items.Task) string {
					if t.Completed {
						return "complete"
					}
					return "reopen"
				},
				"completion",
			)
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
	info := lipgloss.NewStyle().
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat(" ", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// toggleSelectedTask toggles the state of the currently selected task using
// the provided mutation, validation, and labeling functions.
//
// Returns the updated list model and any resulting Bubble Tea commands.
func (m taskPagerModel) toggleSelectedTask(
	toggleFunc func(t *items.Task),
	precondition func(t *items.Task) (bool, string),
	commitKind func(t *items.Task) string,
	actionName string,
) (tea.Model, tea.Cmd) {
	// Clear previous selections.
	for k := range m.listModel.selectedItems {
		delete(m.listModel.selectedItems, k)
	}

	var listModel tea.Model
	var cmds []tea.Cmd

	if selected := m.listModel.list.SelectedItem(); selected != nil {
		t := selected.(*items.Task)

		m.listModel.selectedItems[t.ID] = t

		var toggleCmds []tea.Cmd
		listModel, toggleCmds = m.listModel.toggleTasks(
			toggleFunc,
			precondition,
			commitKind,
			actionName,
		)

		delete(m.listModel.selectedItems, t.ID)
		cmds = append(cmds, toggleCmds...)
	}

	return listModel, tea.Batch(cmds...)
}
