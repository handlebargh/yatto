package models

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
)

type taskFormModel struct {
	form      *huh.Form
	task      *items.Task
	listModel *taskListModel
	width     int
	lg        *lipgloss.Renderer
	styles    *Styles
	vars      *taskFormVars
}

type taskFormVars struct {
	confirm         bool
	taskTitle       string
	taskDescription string
	taskPriority    string
	taskCompleted   bool
}

func newTaskFormModel(t *items.Task, listModel *taskListModel, edit bool) taskFormModel {
	v := taskFormVars{
		confirm:         false,
		taskTitle:       t.Title(),
		taskDescription: t.Description(),
		taskPriority:    t.Priority(),
		taskCompleted:   t.Completed(),
	}

	m := taskFormModel{width: maxWidth}
	m.vars = &v
	m.task = t
	m.listModel = listModel
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var confirmQuestion string
	if edit {
		confirmQuestion = "Edit task?"
	} else {
		confirmQuestion = "Create new task?"
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
				Description("Give it a short but concise title."+"\n"+
					"(max 64 characters)").
				Validate(func(str string) error {
					if len(strings.TrimSpace(str)) < 1 {
						return errors.New("title must not be empty")
					}
					if len(str) > 64 {
						return errors.New("title is too long (only 64 character allowed)")
					}
					return nil
				}),

			huh.NewText().
				Key("description").
				Title("Enter a description:").
				Value(&m.vars.taskDescription),

			huh.NewConfirm().
				Title(confirmQuestion).
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

func (m taskFormModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m taskFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.listModel.spinner, cmd = m.listModel.spinner.Update(msg)
		return m, cmd

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
			m.task.TaskTitle = m.vars.taskTitle
			m.task.TaskDescription = m.vars.taskDescription
			m.task.TaskPriority = m.vars.taskPriority
			m.task.TaskCompleted = m.vars.taskCompleted

			json := items.MarshalTask(
				m.task.Id(),
				m.task.Title(),
				m.task.Description(),
				m.task.Priority(),
				m.task.Completed())

			if storage.FileExists(m.task.Id()) {
				cmds = append(cmds, items.WriteJson(json, *m.task, "update"),
					git.CommitCmd(m.task.Id(),
						"update: "+m.task.Title(),
					),
				)
				m.listModel.loading = true
			} else {
				cmds = append(cmds, items.WriteJson(json, *m.task, "create"),
					git.CommitCmd(m.task.Id(),
						"create: "+m.task.Title(),
					),
				)
				m.listModel.loading = true
			}
		}

		return m.listModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m taskFormModel) View() string {
	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	// Status (right side)
	switch m.vars.taskPriority {
	case "high":
		s.Priority = s.Priority.Background(red)
	case "medium":
		s.Priority = s.Priority.Background(orange)
	case "low":
		s.Priority = s.Priority.Background(indigo)
	default:
		s.Priority = s.Priority.Background(indigo)
	}

	switch m.vars.taskCompleted {
	case true:
		s.Completed = s.Completed.Background(green)
	case false:
		s.Completed = s.Completed.Background(red)
	}

	var status string
	{
		const statusWidth = 35
		statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			Render(s.StatusHeader.Render("Task preview") + "\n\n" +
				s.Title.Render(m.vars.taskTitle) + " " +
				s.Priority.Render(m.vars.taskPriority) + " " +
				s.Completed.Render(items.CompletedString(m.vars.taskCompleted)) + "\n\n" +
				m.vars.taskDescription)
	}

	errors := m.form.Errors()
	header := m.appBoundaryView("Create new task")
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

func (m taskFormModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m taskFormModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m taskFormModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
