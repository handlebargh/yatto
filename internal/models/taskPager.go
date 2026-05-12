// Copyright 2025-2026 handlebargh and contributors
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
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
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
func (m *taskPagerModel) Init() tea.Cmd {
	// Initialize renderer if not already set
	if m.listModel.projectModel.state.renderer == nil {
		isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
		style := "dark"
		if !isDark {
			style = "light"
		}
		renderer, err := glamour.NewTermRenderer(glamour.WithStylePath(style))
		if err != nil {
			// Fallback: store raw content
			m.listModel.projectModel.state.renderer = nil
			m.viewport = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
			m.viewport.SetContent(m.content)
			m.ready = true
			return tea.RequestWindowSize
		}
		m.listModel.projectModel.state.renderer = renderer
	}

	// Render markdown content
	rendered, err := m.listModel.projectModel.state.renderer.Render(m.content)
	if err != nil {
		rendered = m.content // Fallback to raw markdown
	}

	// Initialize viewport with a default size
	// It will be resized when the actual window size message arrives
	m.viewport = viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)
	m.viewport.SetContent(rendered)
	m.ready = true

	return tea.RequestWindowSize
}

// Update handles incoming messages and updates the taskPagerModel accordingly.
func (m *taskPagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Code == 'c' && msg.Mod == tea.ModCtrl {
			return m, tea.Quit
		}

		switch {
		case key.Matches(msg, m.listModel.keys.quit) || key.Matches(msg, m.listModel.keys.goBackVim):
			return m.listModel, nil

		case key.Matches(msg, m.listModel.keys.editItem):
			if m.listModel.list.SelectedItem() != nil {
				// Switch to formModel for editing.
				formModel := newTaskFormModel(m.listModel.list.SelectedItem().(*items.Task), m.listModel, true)
				return formModel, tea.RequestWindowSize
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
			var err error
			// Initialize renderer if not already set
			if m.listModel.projectModel.state.renderer == nil {
				isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
				style := "dark"
				if !isDark {
					style = "light"
				}
				renderer, err := glamour.NewTermRenderer(glamour.WithStylePath(style))
				if err != nil {
					m.listModel.projectModel.state.renderer = nil
				} else {
					m.listModel.projectModel.state.renderer = renderer
				}
			}

			// Render markdown content
			var rendered string
			if m.listModel.projectModel.state.renderer != nil {
				rendered, err = m.listModel.projectModel.state.renderer.Render(m.content)
				if err != nil {
					rendered = "Error rendering markdown"
				}
			} else {
				rendered = m.content // Fallback to raw markdown
			}

			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-footerHeight),
			)
			m.viewport.YPosition = 10
			m.viewport.SetContent(rendered)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - footerHeight)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns the tea.View representation of the task detail view.
func (m *taskPagerModel) View() tea.View {
	if !m.ready {
		content := "\n  Initializing Renderer..."
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}
	content := fmt.Sprintf("%s\n%s", m.viewport.View(), m.footerView())
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// footerView returns the string representation of the task detail view's footer.
func (m *taskPagerModel) footerView() string {
	info := lipgloss.NewStyle().
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat(" ", max(0, m.viewport.Width()-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// toggleSelectedTask toggles the state of the currently selected task using
// the provided mutation, validation, and labeling functions.
//
// Returns the updated list model and any resulting Bubble Tea commands.
func (m *taskPagerModel) toggleSelectedTask(
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
