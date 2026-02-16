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
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/storage"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/muesli/reflow/wordwrap"
)

const (
	// previewWidth defines the width of the task preview box.
	previewWidth = 40

	// previewVerticalPadding positions the status box
	// between header and footer.
	previewVerticalPadding = 10

	// previewContentPadding centers text horizontally inside
	// the staus box.
	previewContentPadding = 3

	// previewLinesToScroll defines how many lines
	// to scroll when pressing pageUP/pageDOWN.
	previewLinesToScroll = 5
)

// taskFormModel defines the Bubble Tea model for a form-based interface
// used to create or edit a task.
type taskFormModel struct {
	form            *huh.Form
	task            *items.Task
	listModel       *taskListModel
	taskLabels      map[string]int
	previewViewport viewport.Model
	userScrolled    bool
	edit            bool
	cancel          bool
	width, height   int
	lg              *lipgloss.Renderer
	styles          *Styles
	vars            *taskFormVars
}

// taskFormVars holds the temporary values that are populated and modified
// in the task form UI.
type taskFormVars struct {
	confirm            bool
	taskTitle          string
	taskDescription    string
	taskPriority       string
	taskDueDate        string
	taskLabels         string
	taskLabelsSelected []string
	taskAuthor         string
	taskAssignee       string
	taskAssigneeNew    string
	taskCompleted      bool
}

// newTaskFormModel initializes and returns a new taskFormModel instance,
// optionally in edit mode.
func newTaskFormModel(t *items.Task, listModel *taskListModel, edit bool) taskFormModel {
	v := taskFormVars{
		confirm:            true,
		taskTitle:          t.Title,
		taskDescription:    t.Description,
		taskPriority:       t.Priority,
		taskDueDate:        t.DueDateToString(),
		taskLabels:         "", // Clear labels as we have them already selected.
		taskLabelsSelected: t.LabelsList(),
		taskAuthor:         t.Author,
		taskAssignee:       t.Assignee,
		taskAssigneeNew:    "", // Clear this field
		taskCompleted:      t.Completed,
	}

	m := taskFormModel{}
	m.edit = edit
	m.vars = &v
	m.task = t
	m.listModel = listModel
	m.taskLabels = helpers.AllLabels(m.listModel.projectModel.config)
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var confirmQuestion string
	if edit {
		if m.vars.taskAuthor == "" {
			m.vars.taskAuthor, _ = vcs.User(m.listModel.projectModel.config)
		}
		confirmQuestion = "Edit task?"
	} else {
		// Ignore error for now
		m.vars.taskAuthor, _ = vcs.User(m.listModel.projectModel.config)
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
				Title("Enter a description:\n"+
					"(markdown is supported)").
				Value(&m.vars.taskDescription),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("dueDate").
				Title(`Valid input formats:

	tomorrow
	next tuesday (or any other weekday)
	in a week
	in a month
	in 3 days
	in 6 weeks

	15:04 (assume today)

	2006-01-28
	2006-01-28 15:04:05

	28.01.2006
	28.01.2006 15:04

	28/01/2006
	28/01/2006 15:04

	Or RFC3339

Date will be in your local timezone
				`).
				Value(&m.vars.taskDueDate).
				Validate(func(str string) error {
					if str == "" {
						return nil
					}

					t, err := parseShortcut(str)
					if err == nil {
						m.vars.taskDueDate = t.Format(time.DateTime)
						return nil
					}

					t, err = parseFlexibleDate(str)
					if err != nil {
						return fmt.Errorf("invalid format")
					}

					if !m.edit && t.Before(time.Now()) {
						return errors.New("due date must be in the future")
					}

					m.vars.taskDueDate = t.Format(time.DateTime)
					return nil
				}),
		).Title("Due Date"),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Key("existingLabels").
				Title("Choose existing labels:").
				Height(15).
				OptionsFunc(m.sortLabelsOptions, nil),

			huh.NewInput().
				Key("labels").
				Title("Enter additional labels:").
				Value(&m.vars.taskLabels).
				Description("Comma-separated list of labels."),
		).Title("Labels"),
		huh.NewGroup(
			huh.NewInput().
				Key("author").
				Title("Enter the task author:").
				Value(&m.vars.taskAuthor).
				Description("This will set the task author."),
		).Title("Author"),
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("existingEmailAddresses").
				Title("Choose an assignee:").
				Height(15).
				OptionsFunc(m.sortEmailAddressesOptions, nil).
				Value(&m.vars.taskAssignee),

			huh.NewInput().
				Key("newEmailAddress").
				Title("Enter a new email address:").
				Value(&m.vars.taskAssigneeNew).
				Description("This will overwrite the selected assignee."),
		).Title("Assignee"),
		huh.NewGroup(
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
func (m taskFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update processes incoming messages and updates the model state accordingly.
func (m taskFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.cancel {
			switch msg.String() {
			case "y", "Y":
				m.cancel = false
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

		switch msg.Type {
		case tea.KeyPgUp:
			m.previewViewport.ScrollUp(previewLinesToScroll)
			m.previewViewport.SetContent(m.generatePreviewContent())
			m.userScrolled = true
		case tea.KeyPgDown:
			m.previewViewport.ScrollDown(previewLinesToScroll)
			m.previewViewport.SetContent(m.generatePreviewContent())

			if m.previewViewport.AtBottom() {
				m.userScrolled = false // Re-enable auto-scroll
			} else {
				m.userScrolled = true
			}
		default:
			break
		}

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v

		m.previewViewport = viewport.New(previewWidth, m.height-previewVerticalPadding)
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)

		// Refresh preview box content.
		m.previewViewport.SetContent(m.generatePreviewContent())

		// Auto-scroll to bottom.
		if !m.userScrolled {
			m.previewViewport.GotoBottom()
		}
	}

	if m.form.State == huh.StateCompleted {
		if m.vars.confirm {
			if m.vars.taskAssigneeNew != "" {
				m.vars.taskAssignee = m.vars.taskAssigneeNew
			}

			err := m.formVarsToTask()
			if err != nil {
				// TODO: we should probably return a message here.
				return m, nil
			}

			json := m.task.MarshalTask()
			taskPath := filepath.Join(m.listModel.project.ID, m.task.ID+".json")

			action := "create"
			if storage.FileExists(m.listModel.projectModel.config, taskPath) {
				action = "update"
			}

			m.listModel.spinning = true
			cmds = append(
				cmds,
				m.listModel.spinner.Tick,
				m.task.WriteTaskJSON(m.listModel.projectModel.config, json, *m.listModel.project, action),
				vcs.CommitCmd(
					m.listModel.projectModel.config,
					fmt.Sprintf("%s: %s", action, m.task.Title),
					taskPath,
				),
			)

			m.listModel.status = ""
			return m.listModel, tea.Batch(cmds...)
		}
		// Return to the start of the form, keep filled in values
		_ = m.formVarsToTask()
		newModel := newTaskFormModel(m.task, m.listModel, m.edit)
		newModel.width = m.width
		newModel.height = m.height
		newModel.previewViewport = viewport.New(previewWidth, m.height-previewVerticalPadding)
		return newModel, newModel.Init()
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
		}

		return centeredStyle.Render("Cancel task creation?\n\n[y] Yes   [n] No")
	}

	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := s.Base.Margin(1, 0).Render(v)

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

	previewMarginLeft := max(0, m.width-previewWidth-lipgloss.Width(form)-s.Status.GetMarginRight())

	status := s.Status.
		MarginLeft(previewMarginLeft).
		BorderForeground(color).
		Width(previewWidth).
		Render(m.previewViewport.View())

	e := m.form.Errors()

	if len(e) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Left, form, status)

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(e) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	var b strings.Builder

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(body)
	b.WriteString("\n\n")
	b.WriteString(footer)

	return s.Base.Render(b.String())
}

// errorView returns a string representation of validation error messages.
func (m taskFormModel) errorView() string {
	var b strings.Builder
	for _, err := range m.form.Errors() {
		b.WriteString(err.Error())
	}
	return b.String()
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

// generatePreviewContent generates the formatted string content for the task preview pane.
// It includes the task title, priority, completion status, and description, all styled
// and wrapped to fit within the width of the preview viewport.
//
// The title line is rendered with appropriate styles for title, priority, and completion status,
// and both the title and description are word-wrapped to avoid overflow.
//
// Returns the full preview string, ready to be set as the viewport's content.
func (m taskFormModel) generatePreviewContent() string {
	s := m.styles

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

	title := fmt.Sprintf("%s %s %s",
		m.styles.Title.Render(m.vars.taskTitle),
		s.Priority.Render(m.vars.taskPriority),
		m.styles.Completed.Render(completedString(m.vars.taskCompleted)))

	var b strings.Builder
	b.WriteString("Task preview")
	b.WriteString("\n\n")
	// We need to wrap our content so it fits into the statusViewport.
	b.WriteString(wordwrap.String(title, previewWidth-previewContentPadding))
	b.WriteString("\n\n")
	b.WriteString(wordwrap.String(m.vars.taskDescription, previewWidth-previewContentPadding))

	// Add due date if set
	if t, err := parseShortcut(m.vars.taskDueDate); err == nil {
		b.WriteString("\n\nDue Date:\n")
		b.WriteString(t.Format(time.RFC1123))
	} else if t, err = parseFlexibleDate(m.vars.taskDueDate); err == nil {
		b.WriteString("\n\nDue Date:\n")
		b.WriteString(t.Format(time.RFC1123))
	}

	return m.styles.StatusHeader.Render(b.String())
}

// formVarsToTask updates the Task object with values from the form variables.
//
// It sets the task's title, description, priority, author, assignee, completion status,
// and due date.
// For labels, it merges labels selected via the multi-select widget with additional
// labels entered as a comma-separated string, deduplicates them (case-insensitive),
// trims whitespace, and stores them as a single comma-separated string on the task.
//
// Returns an error if the due date string cannot be parsed or the local time zone
// cannot be loaded.
func (m taskFormModel) formVarsToTask() error {
	m.task.Title = m.vars.taskTitle
	m.task.Description = m.vars.taskDescription
	m.task.Priority = m.vars.taskPriority
	m.task.Author = m.vars.taskAuthor
	m.task.Assignee = m.vars.taskAssignee

	// Merge labels from MultiSelect (selected) and freeform input (typed)
	typedLabels := helpers.LabelsStringToSlice(m.vars.taskLabels)
	allLabels := append([]string{}, m.form.Get("existingLabels").([]string)...)
	allLabels = append(allLabels, typedLabels...)

	// Deduplicate (case-insensitive) & trim
	labelSet := make(map[string]struct{})
	uniqueLabels := make([]string, 0, len(allLabels))
	for _, label := range allLabels {
		l := strings.TrimSpace(label)
		if l == "" {
			continue // skip empty labels
		}
		key := strings.ToLower(l) // comparison key
		if _, exists := labelSet[key]; !exists {
			labelSet[key] = struct{}{}
			uniqueLabels = append(uniqueLabels, l) // keep original casing
		}
	}

	// Save as comma-separated string
	m.task.Labels = strings.Join(uniqueLabels, ",")

	m.task.Completed = m.vars.taskCompleted

	if m.vars.taskDueDate != "" {
		location, err := time.LoadLocation("Local")
		if err != nil {
			return err
		}

		date, err := time.ParseInLocation(time.DateTime, m.vars.taskDueDate, location)
		if err != nil {
			return err
		}

		m.task.DueDate = &date
	} else {
		m.task.DueDate = nil
	}

	return nil
}

// sortLabelsOptions returns a slice of huh.Option[string] representing the task labels,
// sorted with the following priority:
//  1. Labels currently selected in the form appear first.
//  2. Labels are then sorted by descending frequency (most frequent first).
//  3. Labels with the same frequency are sorted alphabetically (case-insensitive).
//
// The function trims whitespace from label keys and ignores empty labels.
// Selected labels are marked as selected in the returned options.
//
// This method converts the internal map of label frequencies into a sorted slice
// suitable for display in a multi-select UI widget.
func (m taskFormModel) sortLabelsOptions() []huh.Option[string] {
	// Convert map to slice of key-value pairs, trimming whitespace and skipping empties
	type kv struct {
		Label     string
		Frequency int
	}

	var sortedLabels []kv
	for rawLabel, freq := range m.taskLabels {
		label := strings.TrimSpace(rawLabel)
		if label == "" {
			continue // skip empty labels
		}
		sortedLabels = append(sortedLabels, kv{Label: label, Frequency: freq})
	}

	// Convert selected slice to a set
	selectedSet := make(map[string]struct{}, len(m.vars.taskLabelsSelected))
	for _, s := range m.vars.taskLabelsSelected {
		selectedSet[strings.TrimSpace(s)] = struct{}{}
	}

	// Sort: selected first, then by frequency descending, then alphabetical
	slices.SortFunc(sortedLabels, func(a, b kv) int {
		_, aSelected := selectedSet[a.Label]
		_, bSelected := selectedSet[b.Label]

		// Selected labels come first
		if aSelected && !bSelected {
			return -1
		}

		if bSelected && !aSelected {
			return 1
		}

		// Sort by frequency descending
		if a.Frequency != b.Frequency {
			return b.Frequency - a.Frequency
		}

		// Case-insensitive alphabetical
		return strings.Compare(strings.ToLower(a.Label), strings.ToLower(b.Label))
	})

	// Build sorted options
	opts := make([]huh.Option[string], 0, len(sortedLabels))
	for _, item := range sortedLabels {
		opt := huh.NewOption(item.Label, item.Label)
		if _, selected := selectedSet[item.Label]; selected {
			opt = opt.Selected(true)
		}
		opts = append(opts, opt)
	}

	return opts
}

func (m taskFormModel) sortEmailAddressesOptions() []huh.Option[string] {
	emails, _ := vcs.AllContributors(m.listModel.projectModel.config)

	// Sort: selected first, then author's address, then alphabetical
	slices.SortFunc(emails, func(a, b string) int {
		// Selected email comes first
		if a == m.task.Assignee {
			return -1
		}
		if b == m.task.Assignee {
			return 1
		}

		// Author's address
		if a == m.task.Author {
			return -1
		}
		if b == m.task.Author {
			return 1
		}

		// Case-insensitive alphabetical
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	// Build sorted options
	opts := make([]huh.Option[string], 0, len(emails))
	for _, item := range emails {
		opt := huh.NewOption(item, item)
		if item == m.task.Assignee {
			opt = opt.Selected(true)
		}
		opts = append(opts, opt)
	}

	return opts
}

// parseFlexibleDate parses a string into a time.Time value, supporting a variety of common date and time formats.
// It handles ISO 8601, localized formats (e.g., "DD.MM.YYYY", "MM/DD/YYYY"), time-only inputs (assumed for today),
// and RFC3339. If the input does not match any supported format, it returns an error.
//
// Examples of valid inputs:
//   - "2026-02-14"
//   - "14.02.2026 15:04"
//   - "02/14/2026"
//   - "15:04" (assumes today's date)
//   - "2006-01-02T15:04:05Z07:00"
func parseFlexibleDate(str string) (time.Time, error) {
	now := time.Now()
	layouts := []string{
		time.DateTime,      // "2006-01-02 15:04:05"
		"2006-01-02",       // "2006-02-14"
		"02.01.2006 15:04", // "14.02.2026 15:04"
		"02.01.2006",       // "14.02.2026"
		"02/01/2006 15:04", // "02/14/2026 15:04"
		"02/01/2006",       // "02/14/2026"
		"15:04",            // "15:04" (assume today)
		time.RFC3339,       // "2006-01-02T15:04:05Z07:00"
	}

	if len(str) <= 5 {
		if t, err := time.Parse("15:04", str); err == nil {
			return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
		}
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format")
}

// parseShortcut parses natural language date shortcuts into a time.Time value.
// It supports expressions like "tomorrow", "in 3 days", "in 2 weeks", "next monday", etc.
// All returned times are set to midnight (00:00:00) in the local timezone.
// If the input does not match any supported shortcut, it returns an error.
//
// Examples of valid inputs:
//   - "tomorrow"
//   - "in 3 days"
//   - "in 2 weeks"
//   - "next monday"
func parseShortcut(str string) (time.Time, error) {
	now := time.Now()
	todayAtMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	re := regexp.MustCompile(`^in (\d+) (days|weeks)$`)
	matches := re.FindStringSubmatch(strings.ToLower(str))
	if len(matches) == 3 {
		amount, _ := strconv.Atoi(matches[1])
		unit := matches[2]
		switch unit {
		case "days":
			return todayAtMidnight.AddDate(0, 0, amount), nil
		case "weeks":
			return todayAtMidnight.AddDate(0, 0, amount*7), nil
		}
	}

	switch strings.ToLower(str) {
	case "tomorrow":
		return todayAtMidnight.AddDate(0, 0, 1), nil
	case "in a week":
		return todayAtMidnight.AddDate(0, 0, 7), nil
	case "in a month":
		return todayAtMidnight.AddDate(0, 1, 0), nil
	case "next monday":
		return nextWeekday(todayAtMidnight, time.Monday), nil
	case "next tuesday":
		return nextWeekday(todayAtMidnight, time.Tuesday), nil
	case "next wednesday":
		return nextWeekday(todayAtMidnight, time.Wednesday), nil
	case "next thursday":
		return nextWeekday(todayAtMidnight, time.Thursday), nil
	case "next friday":
		return nextWeekday(todayAtMidnight, time.Friday), nil
	case "next saturday":
		return nextWeekday(todayAtMidnight, time.Saturday), nil
	case "next sunday":
		return nextWeekday(todayAtMidnight, time.Sunday), nil
	default:
		return time.Time{}, fmt.Errorf("unknown shortcut")
	}
}

// nextWeekday returns the next occurrence of a specific weekday (e.g., time.Monday)
// after the given time t. If t's weekday is already the target day,
// it returns the time for the same day in the following week.
//
// Example:
//
//	nextWeekday(time.Now(), time.Monday) // Next Monday at the same time as t.
func nextWeekday(t time.Time, day time.Weekday) time.Time {
	diff := (day - t.Weekday() + 7) % 7
	return t.AddDate(0, 0, int(diff))
}
