package models

import (
	"errors"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
)

type projectFormModel struct {
	form          *huh.Form
	project       *items.Project
	listModel     *projectListModel
	edit          bool
	cancel        bool
	width, height int
	lg            *lipgloss.Renderer
	styles        *Styles
	vars          *projectFormVars
}

type projectFormVars struct {
	confirm            bool
	projectTitle       string
	projectDescription string
	projectColor       string
}

func newProjectFormModel(p *items.Project, listModel *projectListModel, edit bool) projectFormModel {
	v := projectFormVars{
		confirm:            false,
		projectTitle:       p.Title(),
		projectDescription: p.Description(),
		projectColor:       p.Color(),
	}

	m := projectFormModel{}
	m.edit = edit
	m.vars = &v
	m.project = p
	m.listModel = listModel
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

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
					if len(str) > 32 {
						return errors.New("title is too long (only 32 character allowed)")
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
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false).
		WithTheme(huh.ThemeBase16())

	// Workaround for a problem that prevents the form
	// from being initially completely rendered.
	m.form.PrevField()

	return m
}

func (m projectFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m projectFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m := newProjectFormModel(m.project, m.listModel, m.edit)
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
			m.project.SetTitle(m.vars.projectTitle)
			m.project.SetDescription(m.vars.projectDescription)
			m.project.SetColor(m.vars.projectColor)

			json := items.MarshalProject(
				m.project.Id(),
				m.project.Title(),
				m.project.Description(),
				m.project.Color(),
			)

			if storage.FileExists(m.project.Id()) {
				cmds = append(cmds,
					m.listModel.progress.SetPercent(0.10),
					tickCmd(),
					items.WriteProjectJson(json, *m.project, "update"),
					git.CommitCmd(filepath.Join(m.project.Id(), "project.json"), "update: "+m.project.Title()),
				)
				m.listModel.status = ""
			} else {
				cmds = append(cmds,
					m.listModel.progress.SetPercent(0.10),
					tickCmd(),
					items.WriteProjectJson(json, *m.project, "create"),
					git.CommitCmd(filepath.Join(m.project.Id(), "project.json"), "create: "+m.project.Title()),
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

func (m projectFormModel) View() string {
	if m.cancel {
		centeredStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

		if m.edit {
			return centeredStyle.Render("Cancel edit?\n\n[y] Yes   [n] No")
		} else {
			return centeredStyle.Render("Cancel project creation?\n\n[y] Yes   [n] No")
		}
	}

	s := m.styles

	// Form
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	var header string
	if m.edit {
		header = m.appBoundaryView("Edit project")
	} else {
		header = m.appBoundaryView("Create new project")
	}

	errors := m.form.Errors()

	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + form + "\n\n" + footer)
}

func (m projectFormModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m projectFormModel) appBoundaryView(text string) string {
	var color lipgloss.AdaptiveColor
	if m.edit {
		color = orange
	} else {
		color = green
	}

	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Foreground(color).Render(text),
		lipgloss.WithWhitespaceChars("❯"),
		lipgloss.WithWhitespaceForeground(color),
	)
}

func (m projectFormModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("❯"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
