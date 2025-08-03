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
	"errors"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
)

// taskFormModel defines the Bubble Tea model for a form-based interface
// used to create or edit a task.
type taskFormModel struct {
	form          *huh.Form
	task          *items.Task
	listModel     *taskListModel
	edit          bool
	cancel        bool
	width, height int
	lg            *lipgloss.Renderer
	styles        *Styles
	vars          *taskFormVars
}

// taskFormVars holds the temporary values that are populated and modified
// in the task form UI.
type taskFormVars struct {
	confirm         bool
	taskTitle       string
	taskDescription string
	taskPriority    string
	taskDueDate     string
	taskLabels      string
	taskCompleted   bool
}

// newTaskFormModel initializes and returns a new taskFormModel instance,
// optionally in edit mode.
func newTaskFormModel(t *items.Task, listModel *taskListModel, edit bool) taskFormModel {
	v := taskFormVars{
		confirm:         true,
		taskTitle:       t.Title(),
		taskDescription: t.Description(),
		taskPriority:    t.Priority(),
		taskDueDate:     t.DueDateToString(),
		taskLabels:      t.Labels(),
		taskCompleted:   t.Completed(),
	}

	m := taskFormModel{}
	m.edit = edit
	m.vars = &v
	m.task = t
	m.listModel = listModel
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var confirmQuestion string
	if edit {
		confirmQuestion = "Edit task?"
	} else {
		confirmQuestion = "Create task?"
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("priority").
				Options(huh.NewOptions("low", "medium", "high")...).
				Title("Select priority").
				Value(&m.vars.taskPriority),

			huh.NewInput().
				Key("title").
				Title("Enter a title:").
				Value(&m.vars.taskTitle).
				Validate(func(str string) error {
					if len(strings.TrimSpace(str)) < 1 {
						return errors.New("title must not be empty")
					}

					return nil
				}),

			huh.NewText().
				Key("description").
				Title("Enter a description:").
				Value(&m.vars.taskDescription),

			huh.NewInput().
				Key("dueDate").
				Title("Enter a due date:").
				Value(&m.vars.taskDueDate).
				Description("Format: YYYY-MM-DD hh:mm:ss").
				Validate(func(str string) error {
					// Ok if no date is set.
					if str == "" {
						return nil
					}

					t, err := time.Parse(time.DateTime, str)
					if err != nil {
						return errors.New("invalid date format")
					}

					if !m.edit && t.Before(time.Now()) {
						return errors.New("date must not be in the past")
					}

					return nil
				}),

			huh.NewInput().
				Key("labels").
				Title("Enter labels:").
				Value(&m.vars.taskLabels).
				Description("Comma separated list of labels."),

			huh.NewConfirm().
				Title(confirmQuestion).
				Affirmative("Yes").
				Negative("No").
				Value(&m.vars.confirm),
		)).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false).
		WithTheme(colors.FormTheme())

	// Workaround for a problem that prevents the form
	// from being initially completely rendered.
	m.form.PrevField()

	return m
}

// Init initializes the form model and returns the initial command to run.
func (m taskFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update processes incoming messages and updates the model state accordingly.
func (m taskFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.cancel {
			switch msg.String() {
			case "y", "Y":
				return m.listModel, nil
			case "n", "N":
				m := newTaskFormModel(m.task, m.listModel, m.edit)
				return m, tea.WindowSize()
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.listModel, nil
		}
	}

	if m.form.State == huh.StateCompleted {
		// Write task only if form has been confirmed.
		if m.vars.confirm {
			m.task.SetTitle(m.vars.taskTitle)
			m.task.SetDescription(m.vars.taskDescription)
			m.task.SetPriority(m.vars.taskPriority)
			m.task.SetLabels(m.vars.taskLabels)
			m.task.SetCompleted(m.vars.taskCompleted)

			if m.vars.taskDueDate != "" {
				// Get the local time zone
				location, err := time.LoadLocation("Local")
				if err != nil {
					// TODO: show an error message
					return m, nil
				}

				date, err := time.ParseInLocation(time.DateTime, m.vars.taskDueDate, location)
				if err != nil {
					// TODO: show an error message
					return m, nil
				}

				m.task.SetDueDate(&date)
			} else {
				m.task.SetDueDate(nil)
			}

			json := m.task.MarshalTask()

			if storage.FileExists(filepath.Join(m.listModel.project.Id(), m.task.Id()+".json")) {
				cmds = append(
					cmds,
					m.listModel.progress.SetPercent(0.10),
					tickCmd(),
					m.task.WriteTaskJson(json, *m.listModel.project, "update"),
					git.CommitCmd(
						filepath.Join(m.listModel.project.Id(), m.task.Id()+".json"),
						"update: "+m.task.Title(),
					),
				)
				m.listModel.status = ""
			} else {
				cmds = append(cmds,
					m.listModel.progress.SetPercent(0.10),
					tickCmd(),
					m.task.WriteTaskJson(json, *m.listModel.project, "create"),
					git.CommitCmd(filepath.Join(m.listModel.project.Id(), m.task.Id()+".json"),
						"create: "+m.task.Title()),
				)
				m.listModel.status = ""
			}
		} else {
			m.cancel = true
			return m, nil
		}

		return m.listModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

// View renders the task form UI and the task preview, depending on the current state.
func (m taskFormModel) View() string {
	if m.cancel {
		centeredStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

		if m.edit {
			return centeredStyle.Render("Cancel edit?\n\n[y] Yes   [n] No")
		} else {
			return centeredStyle.Render("Cancel task creation?\n\n[y] Yes   [n] No")
		}
	}

	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	// Status (right side)
	switch m.vars.taskPriority {
	case "high":
		s.Priority = s.Priority.Background(colors.Red())
	case "medium":
		s.Priority = s.Priority.Background(colors.Orange())
	case "low":
		s.Priority = s.Priority.Background(colors.Indigo())
	default:
		s.Priority = s.Priority.Background(colors.Indigo())
	}

	switch m.vars.taskCompleted {
	case true:
		s.Completed = s.Completed.Background(colors.Green())
	case false:
		s.Completed = s.Completed.Background(colors.Blue())
	}

	var header string
	var color lipgloss.AdaptiveColor
	if m.edit {
		header = m.appBoundaryView("Edit task")
		color = colors.Orange()
	} else {
		header = m.appBoundaryView("Create new task")
		color = colors.Green()
	}

	var status string
	{
		const statusWidth = 40
		statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			BorderForeground(color).
			Render(s.StatusHeader.Render("Task preview") + "\n\n" +
				s.Title.Render(m.vars.taskTitle) + " " +
				s.Priority.Render(m.vars.taskPriority) + " " +
				s.Completed.Render(completedString(m.vars.taskCompleted)) + "\n\n" +
				m.vars.taskDescription)
	}

	errors := m.form.Errors()

	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Left, form, status)

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

// errorView returns a string representation of validation error messages.
func (m taskFormModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

// appBoundaryView returns a formatted header with colored boundaries,
// used for visual separation in the UI.
func (m taskFormModel) appBoundaryView(text string) string {
	var color lipgloss.AdaptiveColor
	if m.edit {
		color = colors.Orange()
	} else {
		color = colors.Green()
	}

	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Foreground(color).Render(text),
		lipgloss.WithWhitespaceChars("❯"),
		lipgloss.WithWhitespaceForeground(color),
	)
}

// appErrorBoundaryView returns a styled horizontal boundary with error-specific colors.
func (m taskFormModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("❯"),
		lipgloss.WithWhitespaceForeground(colors.Red()),
	)
}
