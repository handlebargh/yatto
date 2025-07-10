package models

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/handlebargh/yatto/internal/task"
)

const maxWidth = 80

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Title,
	Priority,
	Completed,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
		Bold(true)
	s.Title = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(green).
		Padding(0, 1).
		Bold(true)
	s.Priority = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
	s.Completed = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type formModel struct {
	form      *huh.Form
	task      *task.Task
	listModel *listModel
	width     int
	lg        *lipgloss.Renderer
	styles    *Styles
}

var (
	confirm         bool = false
	taskTitle       string
	taskDescription string
	taskPriority    string
	taskCompleted   bool
)

func newFormModel(t *task.Task, listModel *listModel, edit bool) formModel {
	if t.Title() != "" {
		taskTitle = t.Title()
	} else {
		taskTitle = "My Task"
	}
	taskDescription = t.Description()
	taskPriority = t.Priority()
	taskCompleted = t.Completed()

	m := formModel{width: maxWidth}
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
				Value(&taskPriority),

			huh.NewInput().
				Key("title").
				Title("Enter a title:").
				Value(&taskTitle).
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
				Value(&taskDescription),

			huh.NewConfirm().
				Title(confirmQuestion).
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		)).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

	// Workaround for a problem that prevents the form
	// from being initially completely rendered.
	m.form.PrevField()

	return m
}

func (m formModel) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, tea.Interrupt
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
		if confirm {
			confirm = false
			m.task.TaskTitle = taskTitle
			m.task.TaskDescription = taskDescription
			m.task.TaskPriority = taskPriority
			m.task.TaskCompleted = taskCompleted

			json := task.MarshalTask(
				m.task.Id(),
				m.task.Title(),
				m.task.Description(),
				m.task.Priority(),
				m.task.Completed())

			if storage.FileExists(m.task.Id()) {
				cmds = append(cmds, task.WriteJsonCmd(json, *m.task, "update: "+m.task.Title()))
				m.listModel.loading = true
			} else {
				m.listModel.list.InsertItem(0, m.task)
				cmds = append(cmds, task.WriteJsonCmd(json, *m.task, "create: "+m.task.Title()))
				m.listModel.loading = true
			}
		}

		return m.listModel, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m formModel) View() string {
	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	// Status (right side)
	switch taskPriority {
	case "high":
		s.Priority = s.Priority.Background(red)
	case "medium":
		s.Priority = s.Priority.Background(orange)
	case "low":
		s.Priority = s.Priority.Background(indigo)
	default:
		s.Priority = s.Priority.Background(indigo)
	}

	switch taskCompleted {
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
				s.Title.Render(taskTitle) + " " +
				s.Priority.Render(taskPriority) + " " +
				s.Completed.Render(task.CompletedString(taskCompleted)) + "\n\n" +
				taskDescription)
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

func (m formModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m formModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m formModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
