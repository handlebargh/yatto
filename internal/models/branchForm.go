package models

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"
)

type branchFormModel struct {
	form      *huh.Form
	branch    *items.Branch
	listModel *branchListModel
	width     int
	lg        *lipgloss.Renderer
	styles    *Styles
	vars      *branchFormVars
}

type branchFormVars struct {
	confirm    bool
	branchName string
	branchType string
}

func newBranchFormModel(b *items.Branch, listModel *branchListModel) branchFormModel {
	v := branchFormVars{
		confirm:    false,
		branchName: b.Title(),
		branchType: b.Description(),
	}

	m := branchFormModel{width: maxWidth}
	m.vars = &v
	m.branch = b
	m.listModel = listModel
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("type").
				Options(huh.NewOptions("local + remote", "local-only")...).
				Title("Select branch type:").
				Description("local-only branches will not leave this computer.").
				Value(&m.vars.branchType),

			huh.NewInput().
				Key("name").
				Title("Enter a name:").
				Description("git branch naming rules apply").
				Value(&m.vars.branchName).
				Validate(func(str string) error {
					if len(strings.TrimSpace(str)) < 1 {
						return errors.New("name must not be empty")
					}
					if len(str) > 32 {
						return errors.New("name is too long (only 32 character allowed)")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Create new branch?").
				Affirmative("Yes").
				Negative("No").
				Value(&m.vars.confirm),
		)).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

	// Workaround for a problem that prevents the form
	// from being initially completely rendered.
	m.form.PrevField()

	return m
}

func (m branchFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m branchFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m.listModel, nil
		}
	}

	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Write task only if form has been confirmed.
		if m.vars.confirm {
			m.vars.confirm = false
			m.branch.Name = m.vars.branchName

			setUpstream := viper.GetBool("git.remote.push_on_commit") && m.vars.branchType == "local + remote"

			cmds = append(cmds,
				m.listModel.progress.SetPercent(0.1),
				tickCmd(),
				git.AddBranchCmd(*m.branch, setUpstream))
		}

		return m.listModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m branchFormModel) View() string {
	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()
	header := m.appBoundaryView("Create new task")
	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + form + "\n\n" + footer)
}

func (m branchFormModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m branchFormModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m branchFormModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
