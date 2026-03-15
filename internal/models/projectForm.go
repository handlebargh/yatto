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
	"errors"
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/mattn/go-runewidth"
)

// projectFormModel defines the Bubble Tea model for a form-based interface
// used to create or edit a project.
type projectFormModel struct {
	form          *huh.Form
	project       *items.Project
	listModel     *ProjectListModel
	edit          bool
	cancel        bool
	width, height int
	styles        lipgloss.Style
	vars          *projectFormVars
}

// projectFormVars holds the temporary values that are populated and modified
// in the project form UI.
type projectFormVars struct {
	confirm            bool
	projectTitle       string
	projectDescription string
	projectColor       string
}

// newProjectFormModel initializes and returns a new projectFormModel instance,
// optionally in edit mode.
func newProjectFormModel(
	p *items.Project,
	listModel *ProjectListModel,
	edit bool,
) projectFormModel {
	v := projectFormVars{
		confirm:            true,
		projectTitle:       p.Title,
		projectDescription: p.Description,
		projectColor:       p.Color,
	}

	m := projectFormModel{}
	m.edit = edit
	m.vars = &v
	m.project = p
	m.listModel = listModel
	m.styles = lipgloss.NewStyle()

	var confirmQuestion string
	if edit {
		confirmQuestion = "Edit project?"
	} else {
		confirmQuestion = "Create new project?"
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("color").
				Options(huh.NewOptions("green", "orange", "red", "blue", "indigo")...).
				Title("Select a color").
				Value(&m.vars.projectColor),

			huh.NewInput().
				Key("title").
				Title("Enter a title:").
				Value(&m.vars.projectTitle).
				Description("Give it a short but concise title."+"\n"+
					"(max 64 characters)").
				Validate(func(str string) error {
					if len(strings.TrimSpace(str)) < 1 {
						return errors.New("title must not be empty")
					}
					if runewidth.StringWidth(str) > 32 {
						return errors.New("title is too long (max 32 terminal columns)")
					}
					return nil
				}),

			huh.NewText().
				Key("description").
				Title("Enter a description:").
				Value(&m.vars.projectDescription),

			huh.NewConfirm().
				Title(confirmQuestion).
				Affirmative("Yes").
				Negative("No").
				Value(&m.vars.confirm),
		)).
		WithWidth(80).
		WithShowHelp(false).
		WithShowErrors(false).
		WithTheme(colors.FormTheme())

	return m
}

// Init initializes the form model and returns the initial command to run.
func (m projectFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update processes incoming messages and updates the model state accordingly.
func (m projectFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.cancel {
			switch msg.String() {
			case "y", "Y":
				return m.listModel, nil
			case "n", "N":
				m.cancel = false
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.cancel = true
			return m, nil
		}

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

	if m.form.State == huh.StateCompleted {
		if !m.vars.confirm {
			return m.listModel, nil
		}

		m.project.Title = m.vars.projectTitle
		m.project.Description = m.vars.projectDescription
		m.project.Color = m.vars.projectColor

		json := m.project.MarshalProject()
		action := "create"
		if storage.FileExists(m.listModel.config, m.project.ID) {
			action = "update"
		}

		m.listModel.spinning = true
		cmds = append(
			cmds,
			m.listModel.spinner.Tick,
			m.project.WriteProjectJSON(m.listModel.config, json, action),
			vcs.CommitCmd(
				m.listModel.config,
				fmt.Sprintf("%s: %s", action, m.project.Title),
				filepath.Join(m.project.ID, "project.json"),
			),
		)

		m.listModel.status = ""
		return m.listModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

// View renders the project form UI.
func (m projectFormModel) View() tea.View {
	var v tea.View
	v.AltScreen = true

	if m.cancel {
		centeredStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

		if m.edit {
			v.SetContent(centeredStyle.Render("Cancel edit?\n\n[y] Yes   [n] No"))

			return v
		}

		v.SetContent(centeredStyle.Render("Cancel project creation?\n\n[y] Yes   [n] No"))

		return v
	}

	// Form
	formView := strings.TrimSuffix(m.form.View(), "\n\n")
	form := lipgloss.NewStyle().Margin(1, 0).Render(formView)

	var header string
	if m.edit {
		header = m.appBoundaryView("Edit project")
	} else {
		header = m.appBoundaryView("Create new project")
	}

	e := m.form.Errors()

	if len(e) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(e) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	var b strings.Builder

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(form)
	b.WriteString("\n\n")
	b.WriteString(footer)

	v.SetContent(b.String())

	return v
}

// errorView returns a string representation of validation error messages.
func (m projectFormModel) errorView() string {
	var b strings.Builder
	for _, err := range m.form.Errors() {
		b.WriteString(err.Error())
	}

	return b.String()
}

// appBoundaryView returns a formatted header with colored boundaries,
// used for visual separation in the UI.
func (m projectFormModel) appBoundaryView(text string) string {
	var whitespaceColor lipgloss.Style
	var color color.Color
	if m.edit {
		whitespaceColor = whitespaceColor.Foreground(colors.Orange())
		color = colors.Orange()
	} else {
		whitespaceColor = whitespaceColor.Foreground(colors.Green())
		color = colors.Green()
	}

	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.Bold(true).Padding(0, 1, 0, 2).Foreground(color).Render(text),
		lipgloss.WithWhitespaceChars("|"),
		lipgloss.WithWhitespaceStyle(whitespaceColor),
	)
}

// appErrorBoundaryView returns a styled horizontal boundary with error-specific colors.
func (m projectFormModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.Foreground(colors.Red()).Render(text),
		lipgloss.WithWhitespaceChars("|"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(colors.Red())),
	)
}
